package process

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"strings"
)

func (proc *Process) String() string                         { return lua.B2S(proc.Byte()) }
func (proc *Process) Type() lua.LValueType                   { return lua.LTObject }
func (proc *Process) AssertFloat64() (float64, bool)         { return 0, false }
func (proc *Process) AssertString() (string, bool)           { return "", false }
func (proc *Process) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (proc *Process) Peek() lua.LValue                       { return proc }

func (proc *Process) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "name":
		return lua.S2L(proc.Name)

	case "pid":
		return lua.LInt(proc.Pid)

	case "ppid":
		return lua.LInt(proc.Ppid)

	case "pgid":
		return lua.LInt(proc.Pgid)

	case "cmd":
		return lua.S2L(proc.Cmdline)

	case "cwd":
		return lua.S2L(proc.Cwd)

	case "exe":
		return lua.S2L(proc.Executable)

	case "state":
		return lua.S2L(proc.State)

	case "args":
		return lua.S2L(strings.Join(proc.Args, " "))

	case "memory":
		return lua.LNumber(proc.MemSize)
	case "rss":
		return lua.LNumber(proc.RssBytes)

	case "rss_pct":
		return lua.LNumber(proc.RssPct)
	case "share":
		return lua.LNumber(proc.Share)

	case "username":
		return lua.S2L(proc.Username)
	case "sha1":
		return lua.S2L(proc.Sha1())
	case "md5":
		return lua.S2L(proc.md5())

	case "stime":
		return lua.S2L(proc.StartTime)
	case "p_cmdline":
		return lua.S2L(proc.ParentCmdline)
	case "p_exe":
		return lua.S2L(proc.ParentExecutable)
	}

	return lua.LNil
}

func CheckById(L *lua.LState, idx int) *Process {
	v := L.Get(idx)
	if v.Type() != lua.LTObject {
		return nil
	}

	p, ok := v.(*Process)
	if ok {
		return p
	}

	return nil
}
