package process

import (
	cond "github.com/vela-ssoc/vela-cond"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"go.uber.org/ratelimit"
	"gopkg.in/tomb.v2"
	"os"
	"time"
)

func (sa *snapshot) deleteL(L *lua.LState) int {
	sa.onDelete.CheckMany(L, pipe.Seek(0))
	return 0
}

func (sa *snapshot) createL(L *lua.LState) int {
	sa.onCreate.CheckMany(L, pipe.Seek(0))
	return 0
}

func (sa *snapshot) updateL(L *lua.LState) int {
	sa.onUpdate.CheckMany(L, pipe.Seek(0))
	return 0
}

/*
func (sa *snapshot) bucketL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	var bkt []string

	for i := 1; i <= n; i++ {
		bkt = append(bkt, L.CheckString(i))
	}

	sa.bkt = bkt
	return 0
}
*/

func (sa *snapshot) runL(L *lua.LState) int {
	sa.V(lua.VTRun, time.Now())
	sa.detect()
	sa.V(lua.VTMode, time.Now())
	return 0
}

func (sa *snapshot) pollL(L *lua.LState) int {
	var interval time.Duration
	n := L.IsInt(1)
	if n < 1 {
		interval = time.Second
	} else {
		interval = time.Duration(n) * time.Second
	}

	sa.tomb = new(tomb.Tomb)
	xEnv.Spawn(0, func() {
		sa.poll(interval)
	})
	sa.V(lua.VTRun, time.Now())
	return 0
}

func (sa *snapshot) limitL(L *lua.LState) int {
	n := L.IsInt(1)
	if n <= 0 {
		return 0
	}

	sa.limit = ratelimit.New(n)
	return 0
}

func (sa *snapshot) ignoreL(L *lua.LState) int {
	if sa.ignore == nil {
		sa.ignore = cond.NewIgnore()
	}

	sa.ignore.CheckMany(L)
	return 0
}

func (sa *snapshot) notAgtL(L *lua.LState) int {
	var exe string
	var err error
	exe, err = os.Executable()
	if err == nil {
		goto done
	}

	exe = os.Args[0]

done:

	if sa.ignore == nil {
		sa.ignore = cond.NewIgnore()
	}
	sa.ignore.Add(cond.New("exe =" + exe))
	sa.ignore.Add(cond.New("p_exe =" + exe))
	sa.ignore.Add(cond.New("name = ssc-worker.exe,ssc-mgt.exe"))
	return 0
}

/*
func (sa *snapshot) pullL(L *lua.LState) int {
	path := L.CheckString(1)
	r := reply{}
	err := xEnv.JSON(path, nil, &r)
	//err := xEnv.GET(path, "").JSON(&r)
	if err != nil {
		L.RaiseError("%s process snap pull process fail %v", sa.Name(), err)
	}

	bkt := xEnv.Bucket(sa.bkt...)
	bkt.Clear()
	sa.debug()

	size := len(r.Data)
	if size == 0 {
		return 0
	}

	tuple := make(map[string]interface{}, size)
	for i := 0; i < size; i++ {
		tuple[auxlib.ToString(r.Data[i].Pid)] = &r.Data[i]
	}
	bkt.BatchStore(tuple, 0)
	return 0
}
*/

func (sa *snapshot) debugL(L *lua.LState) int {
	sa.debug()
	return 0
}

func (sa *snapshot) CacheBucketL(L *lua.LState) int {

	return 0
}

func (sa *snapshot) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	//case "pull":
	//	return lua.NewFunction(sa.pullL)
	case "not_agent":
		return lua.NewFunction(sa.notAgtL)
	case "debug":
		return lua.NewFunction(sa.debugL)
	case "run":
		return lua.NewFunction(sa.runL)
	case "poll":
		return lua.NewFunction(sa.pollL)
	case "limit":
		return lua.NewFunction(sa.limitL)
	case "ignore":
		return lua.NewFunction(sa.ignoreL)
	case "on_create":
		return lua.NewFunction(sa.createL)
	case "on_delete":
		return lua.NewFunction(sa.deleteL)
	case "on_update":
		return lua.NewFunction(sa.updateL)
	case "cache":
		return lua.NewFunction(sa.CacheBucketL)
	case "case":
		return sa.vsh.Index(L, "case")
	}

	return lua.LNil
}
