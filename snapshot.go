package process

import (
	"fmt"
	cond "github.com/vela-ssoc/vela-cond"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"go.uber.org/ratelimit"
	"gopkg.in/tomb.v2"
	"reflect"
	"sync/atomic"
	"time"
)

var (
	snapshotTypeof = reflect.TypeOf((*snapshot)(nil)).String()
)

const (
	SYNC Model = iota + 1
	WORK
)

type Model uint8

func (m Model) String() string {
	switch m {
	case SYNC:
		return "sync"
	case WORK:
		return "work"
	default:
		return ""
	}
}

/*
	664    {"name": "123" , "cwd": "123"}
	665    {"name": "123" , "cwd": "123"}
	667    {"name": "123" , "cwd": "123"}
*/

type reply struct {
	Data []simple `data`
}

type snapshot struct {
	lua.SuperVelaData
	state    uint32
	flag     Model
	co       *lua.LState
	onCreate *pipe.Chains
	onDelete *pipe.Chains
	onUpdate *pipe.Chains
	by       func(int32, *Option) (*Process, error)
	ignore   *cond.Ignore
	tomb     *tomb.Tomb
	limit    ratelimit.Limiter
	bkt      []string
	current  map[int32]*Process
	delete   map[string]interface{}
	update   map[string]*Process
	report   *report

	//report enable
	enable bool

	//report opcode
	opcode int
}

//func FilterSystemProcess(p *Process) bool {
//
//	if p.Username == "root" && p.Executable == "" {
//		return true
//	}
//
//	return false
//}

func newSnapshot(L *lua.LState) *snapshot {
	snt := &snapshot{
		state:    0, //init
		enable:   L.IsTrue(1),
		bkt:      []string{"vela", "process", "snapshot"},
		co:       xEnv.Clone(L),
		onCreate: pipe.New(pipe.Env(xEnv)),
		onDelete: pipe.New(pipe.Env(xEnv)),
		onUpdate: pipe.New(pipe.Env(xEnv)),
	}
	snt.V(lua.VTInit)

	return snt
}

func (snt *snapshot) reset() {
	snt.update = nil
	snt.delete = nil
	snt.report = nil
	snt.current = nil
	snt.flag = 0
}

func (snt *snapshot) withProcess(ps []*Process) {
	n := len(ps)
	if n == 0 {
		return
	}

	if e := xEnv.Push("/api/v1/broker/collect/agent/process/full", ps); e != nil {
		xEnv.Errorf("process snapshot sync push fail %v", e)
	}

	//map fast match
	snt.current = make(map[int32]*Process, n)
	for i := 0; i < n; i++ {
		p := ps[i]
		snt.current[p.Pid] = p
	}

	//by pid find proc
	snt.by = func(pid int32, opt *Option) (*Process, error) {
		proc, ok := snt.current[pid]
		if ok {
			delete(snt.current, pid)
			return proc, nil
		}
		return nil, fmt.Errorf("not found %d process", pid)
	}

}

func (snt *snapshot) Ignore(p *Process) bool {
	if snt.ignore == nil {
		return false
	}

	return snt.ignore.Match(p)
}

func (snt *snapshot) withList(list []int32) {
	n := len(list)
	p := &Process{Pid: -1}
	snt.current = make(map[int32]*Process, n)
	for i := 0; i < n; i++ {
		pid := list[i]
		snt.current[pid] = p
	}

	snt.by = func(pid int32, opt *Option) (*Process, error) {
		delete(snt.current, pid)
		p2, err := Lookup(pid, opt)
		if err != nil {
			return nil, err
		}

		if snt.Ignore(p2) {
			return nil, fmt.Errorf("ignore %v", p2)
		}

		return p2, nil
	}
}

func (snt *snapshot) constructor(flag Model) bool {
	snt.flag = flag
	snt.update = make(map[string]*Process, 128)
	snt.delete = make(map[string]interface{}, 128)
	snt.report = &report{}

	sum := &summary{}
	if sum.init(); !sum.ok() {
		return false
	}
	switch flag {
	case SYNC:
		sum.view(func(p *Process) bool {
			if snt.Ignore(p) {
				return false
			}
			p.Sha1()
			return true
		})
		snt.withProcess(sum.Process)
		return true

	case WORK:
		snt.withList(sum.List())
		return true

	default:
		return false
	}
}

func (snt *snapshot) Name() string {
	return "process.snapshot"
}

func (snt *snapshot) Type() string {
	return snapshotTypeof
}

func (snt *snapshot) Start() error {
	return nil
}

func (snt *snapshot) Close() error {
	if snt.tomb != nil {
		snt.tomb.Kill(nil)
	}

	if snt.limit != nil {
		snt.limit = nil
	}

	return nil
}

func (snt *snapshot) wait() {
	if snt.limit == nil {
		return
	}
	snt.limit.Take()
}

func (snt *snapshot) diff(opt *Option) func(key string, v interface{}) {
	return func(key string, v interface{}) {
		snt.wait()
		pid, err := auxlib.ToInt32E(key)
		if err != nil {
			xEnv.Infof("got invalid pid %v", err)
			snt.delete[key] = v
			snt.report.OnDelete(pid)
			return
		}

		old, ok := v.(*simple)
		if !ok {
			xEnv.Infof("invalid process simple %v", v)
			snt.delete[key] = v
			snt.report.OnDelete(pid)
			return
		}

		if _, exist := snt.current[pid]; !exist {
			snt.delete[key] = v
			snt.report.OnDelete(pid)
			return
		}

		p, er := snt.by(pid, opt)
		if er != nil {
			//xEnv.Infof("not found pid:%d process %v", pid, er)
			snt.delete[key] = v
			snt.report.OnDelete(pid)
			return
		}

		sim := &simple{}
		sim.with(p)
		if !sim.Equal(old) {
			snt.update[key] = p
			p.Sha1()
			p.LookupParent(opt)
			snt.report.OnUpdate(p)
		}

	}
}

func (snt *snapshot) IsRun() bool {
	return atomic.AddUint32(&snt.state, 1) > 1
}

func (snt *snapshot) End() {
	atomic.StoreUint32(&snt.state, 0)
}

func (snt *snapshot) poll(td time.Duration) {
	tk := time.NewTicker(td)
	defer tk.Stop()

	for {
		select {
		case <-snt.tomb.Dying():
			xEnv.Errorf("%s snapshot over", snt.Name())
			return
		case <-tk.C:
			if xEnv.Quiet() {
				continue
			}
			snt.run()
		}
	}
}

func (snt *snapshot) run(opts ...OptionFunc) {
	if snt.IsRun() {
		xEnv.Errorf("process running by %s", snt.flag.String())
		return
	}
	defer snt.End()

	if !snt.constructor(WORK) {
		audit.Errorf("%s process reset snapshot fail", snt.Name()).From(snt.co.CodeVM()).Put()
		return
	}

	opt := NewOption()
	for _, fn := range opts {
		fn(opt)
	}
	opt.Cache = make(map[int32]*Process)

	bkt := xEnv.Bucket(snt.bkt...)
	bkt.Range(snt.diff(opt))
	snt.Create(bkt, opt)
	snt.Delete(bkt, opt)
	snt.Update(bkt, opt)
	snt.doReport()
	snt.reset()
}

func (snt *snapshot) sync(opts ...OptionFunc) {
	if snt.IsRun() {
		xEnv.Errorf("process running by %s", snt.flag.String())
		return
	}
	defer snt.End()

	if !snt.constructor(SYNC) {
		audit.Errorf("%s process reset snapshot fail", snt.Name()).From(snt.co.CodeVM()).Put()
		return
	}

	opt := NewOption()
	for _, fn := range opts {
		fn(opt)
	}

	bkt := xEnv.Bucket(snt.bkt...)
	bkt.Range(snt.diff(opt))
	snt.Create(bkt, opt)
	snt.Delete(bkt, opt)
	snt.Update(bkt, opt)
	snt.reset()
}
