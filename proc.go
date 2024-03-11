package process

import (
	"bytes"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/grep"
	"github.com/vela-ssoc/vela-kit/kind"
	"runtime"
	"time"
)

type Process struct {
	Name       string    `json:"name"`
	Snap       string    `json:"snap"` //快照对比状态
	State      string    `json:"state"`
	Pid        int32     `json:"pid"`
	Ppid       int32     `json:"ppid"`
	Pgid       int32     `json:"pgid"`
	Cmdline    string    `json:"cmdline"`
	Username   string    `json:"username"`
	Cwd        string    `json:"cwd"`
	Executable string    `json:"executable"` // linux
	Checksum   string    `json:"checksum"`
	Md5        string    `json:"md5"`
	Mtime      time.Time `json:"modify_time"`
	Ctime      time.Time `json:"create_time"`
	Args       []string  `json:"args"`
	Uptime     int64     `json:"uptime"`
	CpuPct     float64   `json:"cpu_pct"`
	PidTree    ProcTree  `json:"pid_tree"`

	//Memory
	MemPct float32 `json:"mem_pct"`

	//parent
	ParentName       string `json:"parent_name"`
	ParentCmdline    string `json:"parent_cmdline"`
	ParentExecutable string `json:"parent_executable"`
	ParentUsername   string `json:"parent_username"`
	err              error
	pErr             error
}

func (proc *Process) Ok() bool {
	return proc.err == nil
}

func (proc *Process) Parent() {
	if proc.pErr != nil || proc.err != nil || proc.Ppid == -1 {
		return
	}

	if proc.ParentName != "" {
		return
	}

	ps, err := process.NewProcess(proc.Ppid)
	if err != nil {
		proc.pErr = err
		return
	}

	proc.ParentName, err = ps.Name()
	if err != nil {
		//todo
	}

	exe, err := ps.Exe()
	if err != nil {
		proc.pErr = err
		return
	}
	proc.ParentExecutable = exe

	cmdline, err := ps.Cmdline()
	if err != nil {
		//xEnv.Errorf("pid:%d cmdline got fail %V", ps.Ppid, err)
	} else {
		proc.ParentCmdline = cmdline
	}

	user, err := ps.Username()
	if err != nil {
		//todo
	}
	proc.ParentUsername = user
}

func (proc *Process) ArgToString() string {
	var buf bytes.Buffer

	k := 0
	for _, v := range proc.Args {
		if len(v) == 0 {
			continue
		}

		if k > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(v)
		k++
	}

	return buf.String()
}

func (proc *Process) Byte() []byte {
	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("minion_id", xEnv.ID())
	enc.KV("minion_inet", xEnv.Inet())
	enc.KV("os", runtime.GOOS)
	enc.KV("name", proc.Name)
	enc.KV("state", proc.State)
	enc.KV("pid", proc.Pid)
	enc.KV("ppid", proc.Ppid)
	enc.KV("pgid", proc.Pgid)
	enc.KV("cmdline", proc.Cmdline)
	enc.KV("username", proc.Username)
	enc.KV("cwd", proc.Cwd)
	enc.KV("executable", proc.Executable)
	enc.KV("checksum", proc.Checksum)
	enc.KV("modify_time", proc.Mtime)
	enc.KV("create_time", proc.Ctime)
	enc.KV("snap", proc.Snap)
	enc.KV("uptime", proc.Uptime)
	enc.KV("parent_name", proc.ParentName)
	enc.KV("parent_cmdline", proc.ParentCmdline)
	enc.KV("parent_executable", proc.ParentExecutable)
	enc.KV("parent_username", proc.ParentUsername)
	enc.KV("pid_tree", proc.PidTree.Text())
	enc.End("}")
	return enc.Bytes()
}

func state(b string) string {
	switch b {
	case "S":
		return "sleeping"
	case "R":
		return "running"
	case "D":
		return "idle"
	case "T":
		return "stopped"
	case "Z":
		return "zombie"
	}
	return "unknown"
}

/*

func Fast(pid int32) (*Process, error) {
	proc := &Process{Pid: pid, Snap: "primeval"}
	ps := gosigar.ProcState{}

	err := ps.Get(int(proc.Pid))
	if err != nil {
		return nil, err
	}

	proc.Name = ps.Name
	proc.Ppid = int32(ps.Ppid)
	proc.Pgid = int32(ps.Pgid)
	proc.Username = ps.Username
	proc.State = state(string(ps.State))

	p, err := process.NewProcess(proc.Pid)
	if err != nil {
		xEnv.Errorf("fast process pid:%d name:%s fail %v", proc.Pid, proc.Name, err)
		return nil, err
	}
	proc.LookupExec(p)
	proc.LookupFileStat()
	proc.LookupMem(p)
	proc.LookupCPU(p)
	proc.LookupCreateTime(p)
	return proc, nil
}
*/

func List() []int32 {
	sum := &summary{}
	sum.init()
	return sum.List()
}

//* Name(*)

func Name(pattern string) *summary {
	sum := &summary{}
	sum.name(grep.New(pattern))
	return sum
}

func Lookup(pid int32, opt *Option) (*Process, error) {
	if p := opt.Hit(pid); p != nil {
		return p, nil
	}

	proc := &Process{Pid: pid, Snap: "primeval"}
	err := proc.Lookup(opt)
	if err != nil {
		proc.err = err
		return proc, err
	}

	if opt.Cache != nil {
		opt.Cache[pid] = proc
	}

	return proc, nil
}

func LookupWithBucket(pid int32) *Process {
	key := auxlib.ToString(pid)
	bkt := xEnv.Shm(V_PROC_SHM)

	obj, err := bkt.Get(key)
	if err != nil {
		xEnv.Errorf("not found with bucket %s", key)
		//todo

		proc, _ := Lookup(pid, NewOption())
		return proc
	}

	proc, ok := obj.(*Process)
	if ok {
		return proc
	}

	proc, _ = Lookup(pid, NewOption())
	return proc
}
