package process

import (
	"github.com/shirou/gopsutil/v3/process"
	"strings"
)

func (proc *Process) LookupExec(p *process.Process) error {
	if exe, err := p.Exe(); err == nil {
		proc.Executable = exe
	}

	if args, e := p.CmdlineSlice(); e != nil {
		//xEnv.Errorf("found pid:%d name:%s args fail %v", proc.Pid, proc.Name, e)
	} else {
		proc.Args = args
	}

	if cmd, e := p.Cmdline(); e != nil {
		//xEnv.Errorf("found pid:%d name:%s args fail %v", proc.Pid, proc.Name, e)
	} else {
		proc.Cmdline = cmd
	}

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
		proc.State = strings.Join(v, "|")
	}

	proc.LookupExec(ps)
	proc.LookupMem(ps)
	proc.LookupCPU(ps)
	proc.LookupCreateTime(ps)
	proc.LookupFileStat()
	return nil
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

func Find(pid int32) (*Process, error) {
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
		proc.State = strings.Join(v, "|")
	}

	proc.LookupExec(ps)
	return proc, nil
}

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
		proc.State = strings.Join(v, "|")
	}

	proc.LookupExec(ps)
	proc.LookupCPU(ps)
	proc.LookupCreateTime(ps)
	return proc, nil
}
