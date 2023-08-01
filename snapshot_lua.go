package process

import (
	cond "github.com/vela-ssoc/vela-cond"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"go.uber.org/ratelimit"
	"gopkg.in/tomb.v2"
	"os"
	"time"
)

func (snt *snapshot) deleteL(L *lua.LState) int {
	snt.onDelete.CheckMany(L, pipe.Seek(0))
	return 0
}

func (snt *snapshot) createL(L *lua.LState) int {
	snt.onCreate.CheckMany(L, pipe.Seek(0))
	return 0
}

func (snt *snapshot) updateL(L *lua.LState) int {
	snt.onUpdate.CheckMany(L, pipe.Seek(0))
	return 0
}

func (snt *snapshot) bucketL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	var bkt []string

	for i := 1; i <= n; i++ {
		bkt = append(bkt, L.CheckString(i))
	}

	snt.bkt = bkt
	return 0
}

func (snt *snapshot) runL(L *lua.LState) int {
	snt.V(lua.VTRun, time.Now())
	snt.run()
	snt.V(lua.VTMode, time.Now())
	return 0
}

func (snt *snapshot) pollL(L *lua.LState) int {
	var interval time.Duration
	n := L.IsInt(1)
	if n < 1 {
		interval = time.Second
	} else {
		interval = time.Duration(n) * time.Second
	}

	snt.tomb = new(tomb.Tomb)
	xEnv.Spawn(0, func() {
		snt.poll(interval)
	})
	snt.V(lua.VTRun, time.Now())
	return 0
}

func (snt *snapshot) limitL(L *lua.LState) int {
	n := L.IsInt(1)
	if n <= 0 {
		return 0
	}

	snt.limit = ratelimit.New(n)
	return 0
}

func (snt *snapshot) ignoreL(L *lua.LState) int {
	if snt.ignore == nil {
		snt.ignore = cond.NewIgnore()
	}

	snt.ignore.CheckMany(L)
	return 0
}

func (snt *snapshot) notAgtL(L *lua.LState) int {
	var exe string
	var err error
	exe, err = os.Executable()
	if err == nil {
		goto done
	}

	exe = os.Args[0]

done:

	if snt.ignore == nil {
		snt.ignore = cond.NewIgnore()
	}
	snt.ignore.Add(cond.New("exe =" + exe))
	snt.ignore.Add(cond.New("p_exe =" + exe))
	snt.ignore.Add(cond.New("name = ssc-worker.exe,ssc-mgt.exe"))
	return 0
}

func (snt *snapshot) pullL(L *lua.LState) int {
	path := L.CheckString(1)
	r := reply{}
	err := xEnv.JSON(path, nil, &r)
	//err := xEnv.GET(path, "").JSON(&r)
	if err != nil {
		L.RaiseError("%s process snap pull process fail %v", snt.Name(), err)
	}

	bkt := xEnv.Bucket(snt.bkt...)
	bkt.Clear()
	snt.debug()

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

func (snt *snapshot) debugL(L *lua.LState) int {
	snt.debug()
	return 0
}

func (snt *snapshot) syncL(L *lua.LState) int {
	snt.sync()
	return 0
}

func (snt *snapshot) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "pull":
		return lua.NewFunction(snt.pullL)
	case "not_agent":
		return lua.NewFunction(snt.notAgtL)
	case "debug":
		return lua.NewFunction(snt.debugL)
	case "run":
		return lua.NewFunction(snt.runL)
	case "poll":
		return lua.NewFunction(snt.pollL)
	case "limit":
		return lua.NewFunction(snt.limitL)
	case "ignore":
		return lua.NewFunction(snt.ignoreL)
	case "sync":
		return lua.NewFunction(snt.syncL)
	case "on_create":
		return lua.NewFunction(snt.createL)
	case "on_delete":
		return lua.NewFunction(snt.deleteL)
	case "on_update":
		return lua.NewFunction(snt.updateL)
	}

	return lua.LNil
}
