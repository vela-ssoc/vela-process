package process

import (
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	cond "github.com/vela-ssoc/vela-cond"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"github.com/vela-ssoc/vela-kit/strutil"
	vswitch "github.com/vela-ssoc/vela-switch"
	"go.uber.org/ratelimit"
	"gopkg.in/tomb.v2"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

var (
	snapshotTypeof = reflect.TypeOf((*snapshot)(nil)).String()
)

const (
	V_PROC_SHM = "VELA-PROC-SHM"
)

/*
	664    {"name": "123" , "cwd": "123"}
	665    {"name": "123" , "cwd": "123"}
	667    {"name": "123" , "cwd": "123"}
*/

type ProcEx struct {
	pid   int32
	ps    *process.Process
	value *Process
}

type ProcTree struct {
	Pids []int32
	Tree []string
}

func (pt *ProcTree) Text() string {
	n := len(pt.Tree)
	if n == 0 {
		return ""
	}

	var buf bytes.Buffer
	for i := n - 1; i >= 0; i-- {
		buf.WriteString(strutil.String(pt.Pids[i]))
		buf.WriteString(".")
		buf.WriteString(pt.Tree[i])
		if i != 0 {
			buf.WriteString(">")
		}
	}

	return buf.String()
}

func (pt *ProcTree) Add(pid int32, name string) {
	pt.Pids = append(pt.Pids, pid)
	pt.Tree = append(pt.Tree, name)
}

type snapshot struct {
	lua.SuperVelaData
	state    uint32
	pids     []int32
	co       *lua.LState
	vsh      *vswitch.Switch
	onCreate *pipe.Chains
	onDelete *pipe.Chains
	onUpdate *pipe.Chains
	ignore   *cond.Ignore
	tomb     *tomb.Tomb
	limit    ratelimit.Limiter
	factory  map[int32]*ProcEx
	update   map[int32]*ProcEx
	current  map[int32]bool
	delete   map[string]interface{}
	system   map[int32]string
	report   *report

	//report enable
	enable bool
}

func newSnapshot(L *lua.LState) *snapshot {
	snt := &snapshot{
		state:    0, //init
		enable:   L.IsTrue(1),
		co:       xEnv.Clone(L),
		vsh:      vswitch.NewL(L),
		onCreate: pipe.New(pipe.Env(xEnv)),
		onDelete: pipe.New(pipe.Env(xEnv)),
		onUpdate: pipe.New(pipe.Env(xEnv)),
		limit:    ratelimit.New(150),
	}
	snt.Init(snapshotTypeof)
	return snt
}

func (sa *snapshot) reset() {
	sa.update = nil
	sa.delete = nil
	sa.report = nil
	sa.current = nil
}

func (sa *snapshot) Ignore(p *Process) bool {
	if sa.ignore == nil {
		return false
	}

	return sa.ignore.Match(p)
}

func (sa *snapshot) constructor() error {
	pids, err := process.Pids()
	if err != nil {
		return err
	}

	size := len(pids)
	sa.pids = pids
	sa.factory = make(map[int32]*ProcEx, size)
	sa.update = make(map[int32]*ProcEx, size/2)
	sa.delete = make(map[string]interface{}, size/3)
	sa.system = make(map[int32]string, size/2)
	sa.report = &report{}

	current := make(map[int32]bool, size)
	for i := 0; i < size; i++ {
		pid := pids[i]
		current[pid] = true
	}
	sa.current = current
	return nil
}

func (sa *snapshot) Name() string {
	return "process.snapshot"
}

func (sa *snapshot) Type() string {
	return snapshotTypeof
}

func (sa *snapshot) Start() error {
	return nil
}

func (sa *snapshot) Close() error {
	if sa.tomb != nil {
		sa.tomb.Kill(nil)
	}

	if sa.limit != nil {
		sa.limit = nil
	}

	return nil
}

func (sa *snapshot) wait() {
	if sa.limit == nil {
		return
	}
	sa.limit.Take()
}

func (sa *snapshot) simple(pid int32) (*ProcEx, error) {
	ps, err := process.NewProcess(pid)
	if ps == nil {
		return nil, err
	}

	pv := &Process{Pid: pid}
	if name, e := ps.Name(); e == nil {
		pv.Name = name
	} else {
		return nil, e
	}

	if user, _ := ps.Username(); len(user) > 0 {
		pv.Username = user
	}

	if ppid, _ := ps.Ppid(); ppid > 0 {
		pv.Ppid = ppid
	}

	if stat, _ := ps.Status(); len(stat) > 0 {
		pv.State = strings.Join(stat, "|")
	}

	if uptime, _ := ps.CreateTime(); uptime > 0 {
		pv.Uptime = uptime
	}

	pex := &ProcEx{
		pid:   pid,
		ps:    ps,
		value: pv,
	}

	sa.Factory(pid, pex)

	if name, ok := sa.system[pex.value.Ppid]; ok {
		return nil, fmt.Errorf("ignore system %s process children", name)
	}

	return pex, nil

}

func (sa *snapshot) equal(s *Process, old *Process) bool {
	switch {
	case s.Name != old.Name:
		//xEnv.Errorf("pid=%d s.name = %v  old.name = %v not equal", s.Pid, s, old)
		return false
	case s.Ppid != old.Ppid:
		//xEnv.Errorf("pid=%d s.ppid = %s  old.ppid = %s not equal", s.Pid, s.Ppid, old.Ppid)
		return false
		//case s.Cmdline != old.Cmdline:
		//	//xEnv.Errorf("pid=%d s.cmdline= %s  old.cmdline= %s not equal", s.Pid, s.Cmdline, old.Cmdline)
		//	return false
		//case s.Username != old.Username:
		//	//xEnv.Errorf("pid=%d s.username=%s  old.username= %s not equal", s.Pid, s.Username, old.Username)
		//	return false
		//case s.Executable != old.Executable:
		//	//xEnv.Errorf("pid=%d s.exe= %s  old.exe= %s not equal", s.Pid, s.Executable, old.Executable)
		//return false
	case s.Uptime != old.Uptime:
		xEnv.Errorf("pid=%d s.uptime= %d  old.uptime= %d not equal", s.Pid, s.Uptime, old.Uptime)
		return false
	}
	return true
}

func (sa *snapshot) diff(key string, v interface{}) {

	sa.wait()

	pid, err := auxlib.ToInt32E(key)
	if err != nil {
		xEnv.Infof("got invalid pid %v", err)
		sa.delete[key] = v
		return
	}

	old, ok := v.(*Process)
	if !ok {
		xEnv.Infof("invalid process simple %v", v)
		sa.delete[key] = v
		sa.report.OnDelete(&Process{Pid: pid})
		return
	}

	if _, exist := sa.current[pid]; !exist {
		sa.delete[key] = v
		sa.report.OnDelete(old)
		return
	}

	delete(sa.current, pid)
	pex, er := sa.simple(pid)
	if er != nil {
		sa.delete[key] = v
		sa.report.OnDelete(old)
		return
	}

	if !sa.equal(pex.value, old) {
		sa.update[pid] = pex
	}
}

func (sa *snapshot) Detecting() bool {
	return atomic.AddUint32(&sa.state, 1) > 1
}

func (sa *snapshot) End() {
	atomic.StoreUint32(&sa.state, 0)
}

func (sa *snapshot) poll(td time.Duration) {
	tk := time.NewTicker(td)
	defer tk.Stop()

	for {
		select {
		case <-sa.tomb.Dying():
			xEnv.Errorf("%s snapshot over", sa.Name())
			return
		case <-tk.C:
			if xEnv.Quiet() {
				continue
			}
			sa.detect()
		}
	}
}

func (sa *snapshot) detect() {
	if sa.Detecting() {
		return
	}
	defer sa.End()

	if e := sa.constructor(); e != nil {
		xEnv.Errorf("process snapshot constructor fail %v", e)
		return
	}

	bkt := xEnv.Shm(V_PROC_SHM)
	bkt.Range(sa.diff)
	sa.Create(bkt) //不相等 和 不需要升级 的进程服务
	sa.Update(bkt)
	sa.doReport(bkt)
	sa.Delete(bkt)
	sa.reset()
}
