package process

import (
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/tklauser/go-sysconf"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/strutil"
	"github.com/vela-ssoc/vela-kit/userutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	hertz    = int64(100)
	btime, _ = host.BootTime()
)

func init() {
	if h, err := sysconf.Sysconf(sysconf.SC_CLK_TCK); err == nil {
		hertz = h
	}
}

func (proc *Process) Lookup(opt *Option) error {
	ps, err := process.NewProcess(proc.Pid)
	if err != nil {
		return err
	}

	if e := proc.ReadState(); e != nil {
		return e
	}

	proc.ReadExe()
	proc.ReadStatus()
	proc.ReadCmdline()
	proc.ReadCwd()
	proc.LookupFileStat()
	proc.LookupMem(ps)
	proc.LookupCPU(ps)
	return nil
}

func (proc *Process) ReadState() error {
	var stat []byte
	stat, err := os.ReadFile(filepath.Join("/proc", auxlib.ToString(proc.Pid), "stat"))
	if err != nil {
		return err
	}

	fields := bytes.Fields(stat)
	if len(fields) > 24 {
		proc.Name = string(bytes.TrimFunc(fields[1], func(r rune) bool {
			return r == '(' || r == ')'
		}))
		proc.State = state(string(fields[2]))
		proc.Ppid = auxlib.ToInt32(strutil.B2S(fields[3]))
		proc.Pgid = auxlib.ToInt32(strutil.B2S(fields[4]))

		if uptime, e := strconv.ParseInt(string(fields[21]), 10, 64); e == nil {
			proc.Uptime = uptime/hertz + int64(btime)
		}
	}
	return nil
}

func (proc *Process) ReadCmdline() (cmdline string, err error) {
	if proc.Cmdline != "" {
		cmdline = proc.Cmdline
		return
	}

	var line []byte
	line, err = os.ReadFile(filepath.Join("/proc", auxlib.ToString(proc.Pid), "cmdline"))
	if err != nil {
		return
	}
	cmdline = strutil.B2S(bytes.TrimSpace(bytes.ReplaceAll(line, []byte{0}, []byte{' '})))
	proc.Cmdline = cmdline
	return
}

func (proc *Process) ReadExe() (exe string, err error) {
	if proc.Executable != "" {
		return proc.Executable, nil
	}

	exe, err = os.Readlink(filepath.Join("/proc", auxlib.ToString(proc.Pid), "exe"))
	exe = strings.TrimSpace(exe)
	proc.Executable = exe
	return
}

func (proc *Process) ReadParent() (*Process, error) {
	if proc.Ppid == -1 {
		if e := proc.ReadState(); e != nil {
			return nil, e
		}
	}

	if proc.Ppid == 0 {
		return nil, nil
	}

	ps := &Process{Pid: proc.Ppid}
	if !ps.Have() {
		return nil, fmt.Errorf("not found")
	}

	ps.ReadExe()
	ps.ReadCmdline()
	ps.ReadState()
	ps.ReadStatus()
	proc.ParentName = ps.Name
	proc.ParentUsername = ps.Username
	proc.ParentExecutable = ps.Executable
	proc.ParentCmdline = ps.Cmdline
	return ps, nil
}

func (proc *Process) ReadCwd() (cwd string, err error) {
	if proc.Cwd != "" {
		cwd = proc.Cwd
		return
	}

	cwd, err = os.Readlink(filepath.Join("/proc", auxlib.ToString(proc.Pid), "cwd"))
	proc.Cwd = cwd
	return
}

func (proc *Process) Have() bool {
	_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(int(proc.Pid))))
	if err == nil {
		return true
	}

	return false
}

func (proc *Process) ReadStatus() error {
	var status []byte
	status, err := os.ReadFile(filepath.Join("/proc", auxlib.ToString(proc.Pid), "status"))
	if err != nil {
		return err
	}

	lines := bytes.FieldsFunc(status, func(r rune) bool { return r == '\n' })
	for _, line := range lines {
		fields := bytes.FieldsFunc(line, func(r rune) bool {
			return r == '\t'
		})
		if len(fields) < 2 {
			continue
		}
		key := string(fields[0])

		switch key {
		case "Name:":
			proc.Name = string(fields[1])
		//case "Ppid:":
		//	proc.Ppid = auxlib.ToInt32(strutil.B2S(fields[1]))
		case "Tgid:":
			proc.Pgid = auxlib.ToInt32(strutil.B2S(fields[1]))
		case "Uid:":
			if len(fields) < 5 {
				continue
			} else {
				uid := string(fields[1])
				proc.Username, _ = userutil.Username(uid)
			}
		}
	}

	return nil
}

func Fast(pid int32) (*Process, error) {
	proc := &Process{Pid: pid}
	ps, err := process.NewProcess(pid)
	if err != nil {
		return proc, err
	}

	proc.ReadState()
	proc.ReadStatus()
	proc.ReadCwd()
	proc.ReadCmdline()
	proc.ReadExe()
	proc.LookupCPU(ps)
	proc.LookupCreateTime(ps)
	return proc, nil
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
	proc.ReadState()
	proc.ReadStatus()
	proc.ReadExe()
	proc.ReadCmdline()
	proc.ReadCwd()
	return proc, nil
}

func PsEvent(pid, ppid int32, et string) (*Process, error) {
	ps := &Process{
		Pid:  pid,
		Ppid: ppid,
		Snap: et,
	}

	if !ps.Have() {
		return ps, fmt.Errorf("not found")
	}

	err := ps.ReadState()
	ps.ReadExe()
	ps.ReadCwd()
	ps.ReadCmdline()
	ps.ReadParent()
	ps.ReadStatus()
	ps.Hash()
	return ps, err
}
