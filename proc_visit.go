package process

import (
	"crypto/sha1"
	"fmt"
	"github.com/elastic/gosigar"
	auxlib2 "github.com/vela-ssoc/vela-kit/auxlib"
	"io"
	"os"
	"strings"
)

func (proc *Process) LookupMEM() error {
	mem := gosigar.ProcMem{}
	err := mem.Get(proc.Pid)
	if err != nil {
		return err
	}

	proc.MemSize = mem.Size
	proc.RssBytes = mem.Resident
	proc.Share = mem.Share
	return nil
}

func (proc *Process) LookupCPU() error {
	cpu := gosigar.ProcTime{}
	err := cpu.Get(int(proc.Pid))
	if err != nil {
		return err
	}

	proc.UserTicks = cpu.User
	proc.SystemTicks = cpu.Sys
	proc.TotalTicks = cpu.Total
	return nil
}

func (proc *Process) LookupParent() error {
	exe := gosigar.ProcExe{}
	err := exe.Get(proc.Ppid)
	if err != nil {
		return err
	}

	proc.ParentExecutable = exe.Name

	arg := gosigar.ProcArgs{}
	err = arg.Get(proc.Ppid)
	if err != nil {
		return err
	}
	proc.ParentCmdline = strings.Join(arg.List, " ")

	stat := gosigar.ProcState{}
	err = stat.Get(proc.Ppid)
	if err != nil {
		return err
	}
	proc.ParentUsername = stat.Username
	return nil

}

func (proc *Process) md5() string {
	if proc.Md5 != "" {
		return proc.Md5
	}

	v, _ := auxlib2.FileMd5(proc.Md5)
	proc.Md5 = v
	return v
}

func (proc *Process) Sha1() string {
	if proc.Checksum != "" {
		return proc.Checksum
	}

	if proc.Executable == "" {
		return ""
	}

	fd, err := os.Open(proc.Executable)
	if err != nil {
		return ""
	}
	defer fd.Close()

	info, err := fd.Stat()
	if err != nil {
		return ""
	}
	sh1 := sha1.New()
	io.Copy(sh1, fd)
	proc.Checksum = fmt.Sprintf("%x", sh1.Sum(nil))

	_, mtime, ctime := auxlib2.FileStat(info)
	proc.Mtime = mtime
	proc.Ctime = ctime

	return proc.Checksum
}

func (proc *Process) LookupState() error {
	st := gosigar.ProcState{}

	err := st.Get(proc.Pid)
	if err != nil {
		return err
	}

	proc.Name = st.Name
	proc.State = state(byte(st.State))
	proc.Ppid = st.Ppid
	proc.Pgid = uint32(st.Pgid)
	proc.Username = st.Username
	proc.LookupExec()
	proc.LookupMEM()
	proc.LookupCPU()
	proc.LookupParent()
	return nil
}

func (proc *Process) Lookup() error {
	return proc.LookupState()
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
