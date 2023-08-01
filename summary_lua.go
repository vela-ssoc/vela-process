package process

import (
	"encoding/json"
	"github.com/vela-ssoc/vela-kit/grep"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
)

func (sum *summary) String() string                         { return lua.B2S(sum.Byte()) }
func (sum *summary) Type() lua.LValueType                   { return lua.LTObject }
func (sum *summary) AssertFloat64() (float64, bool)         { return 0, false }
func (sum *summary) AssertString() (string, bool)           { return "", false }
func (sum *summary) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (sum *summary) Peek() lua.LValue                       { return sum }

func (sum *summary) Byte() []byte {
	chunk, _ := json.Marshal(sum)
	return chunk
}

func (sum *summary) showL(L *lua.LState) int {
	if L.Console == nil {
		return 0
	}

	var i uint32 = 0

	for ; i < sum.Total; i++ {
		pv := sum.Process[i]
		L.Output(pv.String())
	}

	return 0
}

func (sum *summary) pipeL(L *lua.LState) int {
	filter := fuzzy(grep.New(L.IsString(1)))
	pp := pipe.NewByLua(L, pipe.Seek(1))
	var i uint32 = 0

	for ; i < sum.Total; i++ {
		pv := sum.Process[i]
		if !filter(pv) {
			continue
		}
		pp.Do(pv, L, func(err error) {
			xEnv.Errorf("sum process pipe fail %v", err)
		})
	}
	return 0
}

func (sum *summary) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	switch key.Type() {
	case lua.LTString:
		return sum.Index(L, key.String())
	case lua.LTInt:
		return sum.GetByIndex(int(key.(lua.LInt)))
	}

	return lua.LNil

}

func (sum *summary) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "total":
		return lua.LInt(sum.Total)
	case "case":
		return sum.vsh.Index(L, "key")

	case "run":
		return lua.LInt(sum.Running)

	case "sleep":
		return lua.LInt(sum.Sleeping)

	case "stop":
		return lua.LInt(sum.Stopped)

	case "idle":
		return lua.LInt(sum.Idle)

	case "pipe":
		return lua.NewFunction(sum.pipeL)

	case "show":
		return lua.NewFunction(sum.showL)

	}

	return lua.LNil
}
