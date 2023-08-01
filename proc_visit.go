package process

import (
	"github.com/elastic/gosigar"
	"github.com/shirou/gopsutil/process"
	"github.com/vela-ssoc/vela-kit/fileutil"
	"strings"
)

func (proc *Process) Kill() error {
	p, err := process.NewProcess(int32(proc.Pid))
	if err != nil {
		return err
	}
	return p.Kill()
}

func (proc *Process) OpenFiles() ([]process.OpenFilesStat, error) {
	u, err := process.NewProcess(int32(proc.Pid))
	if err != nil {
		return nil, err
	}

	return u.OpenFiles()
}

func (proc *Process) LookupCPU(p *process.Process) error {
	cpu := gosigar.ProcTime{}
	err := cpu.Get(int(proc.Pid))
	if err != nil {
		return err
	}

	proc.UserTicks = cpu.User
	proc.SystemTicks = cpu.Sys
	proc.TotalTicks = cpu.Total

	if pct, e := p.CPUPercent(); e == nil {
		proc.CpuPct = pct
	} else {
		xEnv.Infof("pid:%d name:%s cpu percent fail %v", proc.Pid, proc.Name, e)
	}

	return nil
}

func (proc *Process) LookupParent(opt *Option) error {
	if pp := opt.Hit(proc.Ppid); pp != nil {
		proc.ParentExecutable = pp.Executable
		proc.ParentUsername = pp.Username
		proc.ParentCmdline = pp.Cmdline
		return nil
	}

	pp, err := Lookup(proc.Ppid, opt)
	if err != nil {
		return err
	}

	proc.ParentExecutable = pp.Executable
	proc.ParentCmdline = pp.Cmdline
	proc.ParentUsername = pp.Username
	return nil
}

func (proc *Process) md5() string {
	if proc.Md5 != "" {
		return proc.Md5
	}

	if proc.Executable == "" {
		return ""
	}

	csm, err := hash(proc.Executable)
	if err != nil {
		return ""
	}
	proc.Md5 = csm.Md5
	proc.Checksum = csm.Sha1
	return proc.Md5
}

func (proc *Process) Sha1() string {
	if proc.Checksum != "" {
		return proc.Checksum
	}

	if proc.Executable == "" {
		return ""
	}

	csm, err := hash(proc.Executable)
	if err == nil {
		proc.Checksum = csm.Sha1
		proc.Md5 = csm.Md5
	}

	return proc.Checksum
}

func (proc *Process) LookupMem(p *process.Process) error {
	mem, err := p.MemoryInfo()
	if err != nil {
		return err
	}
	proc.MemSize = mem.HWM
	proc.RssBytes = mem.RSS
	proc.Share = mem.Swap

	if pct, e := p.MemoryPercent(); e == nil {
		proc.MemPct = pct
	} else {
		xEnv.Infof("pid:%d name:%s memory percent fail %v", proc.Pid, proc.Name, e)
	}

	return nil
}

func (proc *Process) LookupCreateTime(p *process.Process) error {
	ctime, err := p.CreateTime()
	if err != nil {
		return err
	}

	proc.Uptime = ctime
	return nil
}

func (proc *Process) LookupFileStat() error {
	if len(proc.Executable) == 0 {
		return nil
	}

	_, mt, ct, _, err := fileutil.State(proc.Executable)
	if err != nil {
		return err
	}

	proc.Ctime = ct
	proc.Mtime = mt

	return nil
}

func (proc *Process) Lookup(opt *Option) error {
	ps, err := process.NewProcess(proc.Pid)
	if err != nil {
		return err
	}

	if v, e := ps.Name(); e == nil {
		proc.Name = v
	} else {
		return e
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

	proc.LookupExec(ps)
	proc.LookupMem(ps)
	proc.LookupCPU(ps)
	proc.LookupCreateTime(ps)
	proc.LookupFileStat()
	return nil
}

func (proc *Process) IsNull() bool {
	return proc == nil || proc.Pid == -1
}

func (proc *Process) ArgsToString() string {
	if len(proc.Args) == 0 {
		return ""
	}
	return strings.Join(proc.Args, " ")
}
