package process

import (
	"github.com/vela-ssoc/vela-kit/lua"
)

func (s *simple) String() string                         { return lua.B2S(s.Byte()) }
func (s *simple) Type() lua.LValueType                   { return lua.LTObject }
func (s *simple) AssertFloat64() (float64, bool)         { return 0, false }
func (s *simple) AssertString() (string, bool)           { return "", false }
func (s *simple) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (s *simple) Peek() lua.LValue                       { return s }

func (s *simple) Byte() []byte {

	p := Process(*s)
	return p.Byte()

	/*
		enc := kind.NewJsonEncoder()

		enc.Tab("")
		enc.KV("minion_id", xEnv.ID())
		enc.KV("minion_inet", xEnv.Inet())
		enc.KV("name", s.Name)
		enc.KV("state", s.State)
		enc.KV("pid", s.Pid)
		enc.KV("ppid", s.p)
		enc.KV("pgid", s.PGid)
		enc.KV("cmdline", s.Cmdline)
		enc.KV("username", s.Username)
		enc.KV("cwd", s.Cwd)
		enc.KV("executable", s.Executable)
		enc.KV("args", strings.Join(s.Args, " "))
		enc.End("}")

		return enc.Bytes()
	*/
}

func (s *simple) Index(L *lua.LState, key string) lua.LValue {
	p := Process(*s)
	return p.Index(L, key)

	/*
		switch key {
		case "name":
			return lua.S2L(s.Name)
		case "state":
			return lua.S2L(s.State)
		case "pid":
			return lua.LInt(s.Pid)
		case "ppid":
			return lua.LInt(s.PPid)
		case "pgid":
			return lua.LInt(s.PGid)
		case "cmdline":
			return lua.S2L(s.Cmdline)
		case "username":
			return lua.S2L(s.Username)
		case "cwd":
			return lua.S2L(s.Cwd)
		case "executable":
			return lua.S2L(s.exe())
		case "args":
			return lua.S2L(strings.Join(s.Args, " "))
		}

		return lua.LNil

	*/
}
