package process

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"runtime"
)

func (proc *Process) String() string                         { return lua.B2S(proc.Byte()) }
func (proc *Process) Type() lua.LValueType                   { return lua.LTObject }
func (proc *Process) AssertFloat64() (float64, bool)         { return 0, false }
func (proc *Process) AssertString() (string, bool)           { return "", false }
func (proc *Process) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (proc *Process) Peek() lua.LValue                       { return proc }

/*
vela.ps().pipe(function(p)
	p.handles
end)

*/

func (proc *Process) Handle() lua.LValue {
	s, err := proc.OpenFiles()
	if err != nil {
		xEnv.Debugf("pid=%d handle check fail %v", proc.Pid, err)
		return &HandleSummary{Pid: proc.Pid, Err: err, Files: s}
	}

	return &HandleSummary{Pid: proc.Pid, Err: err, Files: s}
}

func (proc *Process) showL(L *lua.LState) int {
	if L.Console == nil {
		return 0
	}

	L.Output(proc.String())
	return 0
}

func (proc *Process) params() lua.LValue {
	params := lua.NewSlice(0)

	n := len(proc.Args)
	if n == 0 {
		return params
	}

	for i := 0; i < n; i++ {
		if v := proc.Args[i]; len(v) > 0 {
			params = append(params, lua.S2L(v))
		}
	}
	return params
}

func (proc *Process) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "os":
		return lua.S2L(runtime.GOOS)
	case "minion_ip", "ip":
		return lua.S2L(xEnv.Inet())
	case "minion_id":
		return lua.S2L(xEnv.ID())
	case "ok":
		return lua.LBool(proc.err == nil)

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

	case "cmdline":
		return lua.S2L(proc.Cmdline)

	case "cwd":
		return lua.S2L(proc.Cwd)

	case "exe":
		return lua.S2L(proc.Executable)

	case "state":
		return lua.S2L(proc.State)

	case "args":
		return proc.params()

	case "username":
		return lua.S2L(proc.Username)
	case "sha1":
		return lua.S2L(proc.Sha1())
	case "md5":
		return lua.S2L(proc.md5())
	case "p_name":
		proc.Parent()
		if proc.pErr == nil {
			return lua.S2L(proc.ParentName)
		}
		return lua.LSNull

	case "p_cmdline":
		proc.Parent()
		if proc.pErr == nil {
			return lua.S2L(proc.ParentCmdline)
		}
		return lua.LSNull
	case "p_exe":
		proc.Parent()
		if proc.pErr == nil {
			return lua.S2L(proc.ParentExecutable)
		}
		return lua.LSNull
	case "p_username":
		proc.Parent()
		if proc.pErr == nil {
			return lua.S2L(proc.ParentUsername)
		}
		return lua.LSNull

	case "handle":
		return proc.Handle()
	case "uptime":
		return lua.LInt64(proc.Uptime)
	case "show":
		return lua.NewFunction(proc.showL)
	case "pid_tree":
		return lua.S2L(proc.PidTree.Text())
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
