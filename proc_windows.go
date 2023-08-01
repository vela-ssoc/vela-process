package process

import (
	"github.com/shirou/gopsutil/process"
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
