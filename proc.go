package process

import (
	"bytes"
	"github.com/vela-ssoc/vela-kit/grep"
	"github.com/vela-ssoc/vela-kit/kind"
	"time"
)

type Process struct {
	Name       string    `json:"name"`
	Snap       string    `json:"snap"` //快照对比状态
	State      string    `json:"state"`
	Pid        int       `json:"pid"`
	Ppid       int       `json:"ppid"`
	Pgid       uint32    `json:"pgid"`
	Cmdline    string    `json:"cmdline"`
	Username   string    `json:"username"`
	Cwd        string    `json:"cwd"`
	Executable string    `json:"executable"` // linux
	Checksum   string    `json:"checksum"`
	Md5        string    `json:"md5"`
	Mtime      time.Time `json:"modify_time"`
	Ctime      time.Time `json:"create_time"`
	Args       []string  `json:"args"`

	//CPU，单位 毫秒
	UserTicks    uint64  `json:"user_ticks"`
	TotalPct     float64 `json:"total_pct"`
	TotalNormPct float64 `json:"total_norm_pct"`
	SystemTicks  uint64  `json:"system_ticks"`
	TotalTicks   uint64  `json:"total_ticks"`
	StartTime    string  `json:"start_time"`

	//Memory
	MemSize  uint64  `json:"mem_size"`
	RssBytes uint64  `json:"rss_bytes"`
	RssPct   float64 `json:"rss_pct"`
	Share    uint64  `json:"share"`

	//parent
	ParentCmdline    string `json:"parent_cmdline"`
	ParentExecutable string `json:"parent_executable"`
	ParentUsername   string `json:"parent_username"`

	err error
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
	enc.KV("args", proc.ArgToString())
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
	enc.KV("parent_cmdline", proc.ParentCmdline)
	enc.KV("parent_executable", proc.ParentExecutable)
	enc.KV("parent_username", proc.ParentUsername)
	enc.End("}")
	return enc.Bytes()
}

func state(b byte) string {
	switch b {
	case 'S':
		return "sleeping"
	case 'R':
		return "running"
	case 'D':
		return "idle"
	case 'T':
		return "stopped"
	case 'Z':
		return "zombie"
	}
	return "unknown"
}

func Pid(pid int) (*Process, error) {
	proc := &Process{Pid: pid, Snap: "primeval"}
	err := proc.Lookup()
	return proc, err
}

func List() []int {
	sum := &summary{}
	sum.init()
	return sum.List()
}

func Name(pattern string) *summary {
	sum := &summary{}
	sum.name(grep.New(pattern))
	return sum
}
