package process

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"github.com/vela-ssoc/vela-kit/execpt"
)

func (proc *Process) LookupExec() error {
	p, err := process.NewProcess(int32(proc.Pid))
	if err != nil {
		return err
	}

	c := execpt.New()
	s := fmt.Sprintf

	exe, err := p.Exe()
	c.Try(s("%d exe", proc.Pid), err)
	proc.Executable = exe

	return c.Wrap()
}
