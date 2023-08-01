package process

import (
	"bytes"
	"github.com/shirou/gopsutil/process"
	"github.com/vela-ssoc/vela-kit/grep"
	"github.com/vela-ssoc/vela-kit/kind"
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
	//CPU，单位 毫秒
	UserTicks    uint64  `json:"user_ticks"`
	TotalPct     float64 `json:"total_pct"`
	TotalNormPct float64 `json:"total_norm_pct"`
	SystemTicks  uint64  `json:"system_ticks"`
	TotalTicks   uint64  `json:"total_ticks"`
	StartTime    string  `json:"start_time"`
	CpuPct       float64 `json:"cpu_pct"`

	//Memory
	MemSize  uint64  `json:"mem_size"`
	RssBytes uint64  `json:"rss_bytes"`
	RssPct   float64 `json:"rss_pct"`
	Share    uint64  `json:"share"`
	MemPct   float32 `json:"mem_pct"`

	//parent
	ParentCmdline    string `json:"parent_cmdline"`
	ParentExecutable string `json:"parent_executable"`
	ParentUsername   string `json:"parent_username"`

	err  error
	pErr error
}

func (proc *Process) Parent() {
	if proc.pErr != nil {
		return
	}

	if proc.ParentExecutable != "" {
		return
	}

	ps, err := process.NewProcess(proc.Ppid)
	if err != nil {
		proc.pErr = err
		return
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
		proc.Cmdline = cmdline
	}

	user, err := ps.Username()
	if err != nil {
		//todo
	}
	proc.Username = user
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
	enc.KV("name", proc.Name)
	enc.KV("state", proc.State)
	enc.KV("pid", proc.Pid)
	enc.KV("ppid", proc.Ppid)
	enc.KV("pgid", proc.Pgid)
	enc.KV("cmdline", proc.ArgsToString())
	enc.KV("username", proc.Username)
	enc.KV("cwd", proc.Cwd)
	enc.KV("executable", proc.Executable)
	enc.KV("args", proc.Args)
	enc.KV("checksum", proc.Checksum)
	enc.KV("modify_time", proc.Mtime)
	enc.KV("create_time", proc.Ctime)

	enc.KV("user_ticks", proc.UserTicks)
	enc.KV("total_pct", proc.TotalPct)
	enc.KV("total_norm_pct", proc.TotalNormPct)
	enc.KV("system_ticks", proc.SystemTicks)
	enc.KV("total_ticks", proc.TotalTicks)
	enc.KV("start_time", proc.StartTime)

	enc.KV("mem_size", proc.MemSize)
	enc.KV("rss_bytes", proc.RssBytes)
	enc.KV("rss_pct", proc.RssPct)
	enc.KV("share", proc.Share)
	enc.KV("snap", proc.Snap)
	enc.KV("uptime", proc.Uptime)
	enc.KV("parent_cmdline", proc.ParentCmdline)
	enc.KV("parent_executable", proc.ParentExecutable)
	enc.KV("parent_username", proc.ParentUsername)
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

func Pid(pid int32, opts ...OptionFunc) (*Process, error) {
	opt := &Option{
		Cpu:    true,
		Mem:    true,
		Parent: true,
	}

	for _, fn := range opts {
		fn(opt)
	}
	return Lookup(pid, opt)
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

func Fast(pid int32) (*Process, error) {
	proc := &Process{Pid: pid}
	ps, err := process.NewProcess(pid)
	if err != nil {
		return proc, err
	}

	if v, e := ps.Name(); e == nil {
		proc.Name = v
	} else {
		return proc, e
	}

	if v, e := ps.Ppid(); e == nil {
		proc.Ppid = v
	}

	if v, e := ps.Tgid(); e == nil {
		proc.Pgid = v
	}

	if v, e := ps.Username(); e == nil {
		proc.Username = v
	}

	if v, e := ps.Status(); e == nil {
		proc.State = state(v)
	}
	proc.LookupCPU(ps)
	proc.LookupCreateTime(ps)
	return proc, nil
}

func List() []int32 {
	sum := &summary{}
	sum.init()
	return sum.List()
}

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
		return nil, err
	}

	if opt.Cache != nil {
		opt.Cache[pid] = proc
	}

	return proc, nil
}
