//go:build linux || darwin
// +build linux darwin

package process

import (
	"github.com/shirou/gopsutil/process"
)

func (proc *Process) LookupExec(p *process.Process) error {
	if exe, err := p.Exe(); err == nil {
		proc.Executable = exe
	} else {
		//xEnv.Infof("pid:%d name:%s exe fail %v", proc.Pid, proc.Name, err)
	}

	if cwd, e := p.Cwd(); e != nil {
		//xEnv.Errorf("found pid:%d name:%s cwd fail %v", p.Pid, name, err)
	} else {
		proc.Cwd = cwd
	}

	if args, e := p.CmdlineSlice(); e != nil {
		xEnv.Infof("found pid:%d name:%s args fail %v", p.Pid, proc.Name, e)
	} else {
		proc.Args = args
	}

	if cmd, e := p.Cmdline(); e != nil {
		xEnv.Errorf("found pid:%d name:%s args fail %v", p.Pid, proc.Name, e)
	} else {
		proc.Cmdline = cmd
	}
	return nil
}
