package process

import (
	"github.com/vela-ssoc/vela-kit/vela"
	cond "github.com/vela-ssoc/vela-cond"
	"github.com/vela-ssoc/vela-kit/grep"
	"github.com/vela-ssoc/vela-kit/lua"
)

var xEnv vela.Environment

func pidL(L *lua.LState) int {
	pid := L.IsInt(1)
	if pid == 0 {
		return 0
	}

	proc, err := Pid(pid)
	if err != nil {
		return 0
	}

	L.Push(proc)
	return 1
}

func nameL(L *lua.LState) int {
	sum := NewSumL(L)
	sum.name(grep.New(L.IsString(1)))
	L.Push(sum)
	return 1
}

func cmdL(L *lua.LState) int {
	sum := NewSumL(L)
	sum.cmd(grep.New(L.IsString(1)))
	L.Push(sum)
	return 1
}
func exeL(L *lua.LState) int {
	sum := NewSumL(L)
	sum.exe(grep.New(L.IsString(1)))
	L.Push(sum)
	return 1
}

func userL(L *lua.LState) int {
	sum := NewSumL(L)
	sum.user(grep.New(L.IsString(1)))
	L.Push(sum)
	return 1
}

func cwdL(L *lua.LState) int {
	sum := NewSumL(L)
	sum.cwd(grep.New(L.IsString(1)))
	L.Push(sum)
	return 1
}

func ppidL(L *lua.LState) int {
	sum := NewSumL(L)
	sum.ppid(grep.New(L.IsString(1)))
	L.Push(sum)
	return 1
}

func allL(L *lua.LState) int {
	sum := NewSumL(L)
	cnd := cond.CheckMany(L, cond.Seek(0))
	sum.init()
	if !sum.ok() {
		goto done
	}

	sum.search(cnd)

done:
	L.Push(sum)
	return 1
}

func snapshotL(L *lua.LState) int {
	snt := newSnapshot(L)
	if snt == nil {
		L.RaiseError("new process snapshot fail")
		return 0
	}

	proc := L.NewVelaData(snt.Name(), snt.Type())
	if proc.IsNil() {
		proc.Set(snt)
	} else {
		old := proc.Data.(*snapshot)
		old.Close()
		proc.Set(snt)
	}

	L.Push(proc)
	return 1
}

/*
	local sum = rock.ps.all()
	local sum = rock.ps.name("*dlv*")

	sum.pipe(_(p)
		p.name
		p.cmd
		p.cwd
		p.exe
		p.ppid
	end)


	local p = rock.ps.pid(123)

	local snap = rock.ps.snapshot()

	snap.poll(5)
*/

func WithEnv(env vela.Environment) {
	xEnv = env
	tab := lua.NewUserKV()
	tab.Set("pid", lua.NewFunction(pidL))
	tab.Set("exe", lua.NewFunction(exeL))
	tab.Set("cmd", lua.NewFunction(cmdL))
	tab.Set("user", lua.NewFunction(userL))
	tab.Set("cwd", lua.NewFunction(cwdL))
	tab.Set("name", lua.NewFunction(nameL))
	tab.Set("ppid", lua.NewFunction(ppidL))
	tab.Set("snapshot", lua.NewFunction(snapshotL))

	env.Set("ps",
		lua.NewExport("vela.ps.export",
			lua.WithTable(tab),
			lua.WithFunc(allL)))

	//注册加解密
	xEnv.Mime(simple{}, encode, decode)
}
