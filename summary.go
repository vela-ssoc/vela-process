package process

import (
	"github.com/shirou/gopsutil/v3/process"
	cond "github.com/vela-ssoc/vela-cond"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	vswitch "github.com/vela-ssoc/vela-switch"
)

type summary struct {
	Idle     uint32 `json:"idle"`
	Running  uint32 `json:"running"`
	Sleeping uint32 `json:"sleeping"`
	Stopped  uint32 `json:"stopped"`
	Total    uint32 `json:"total"`
	Unknown  uint32 `json:"unknown"`
	Zombie   uint32 `json:"zombie"`

	Process []*Process      `json:"process"`
	Pids    []int32         `json:"Pids"`
	Error   error           `json:"-"`
	vsh     *vswitch.Switch `json:"-"`
	co      *lua.LState     `json:"-"`
}

func (sum *summary) List() []int32 {
	return sum.Pids
}

func (sum *summary) Map() map[int32]*Process {
	if n := len(sum.Process); n != 0 {
		tab := make(map[int32]*Process, n)
		for i := 0; i < n; i++ {
			p := sum.Process[i]
			tab[p.Pid] = sum.Process[i]
		}
		return tab
	}

	n := len(sum.Pids)
	p := &Process{Pid: -1}
	tab := make(map[int32]*Process, n)
	for i := 0; i < n; i++ {
		pid := sum.Pids[i]
		tab[pid] = p
	}
	return tab
}

func (sum *summary) append(pv *Process) {
	switch pv.State {
	case "sleeping":
		sum.Sleeping++
	case "running":
		sum.Running++
	case "idle":
		sum.Idle++
	case "stopped":
		sum.Stopped++
	case "zombie":
		sum.Zombie++
	default:
		sum.Unknown++
	}
	sum.Total++
	sum.Process = append(sum.Process, pv)
}

func (sum *summary) init() {
	pids, err := process.Pids()
	if err != nil {
		sum.Error = err
		return
	}
	sum.Pids = pids
}

func (sum *summary) ok() bool {
	return sum.Error == nil
}

func (sum *summary) name(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		return match(pv.Name)
	})

	return
}

func (sum *summary) cmd(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		return match(pv.Cmdline)
	})
	return
}

func (sum *summary) exe(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		return match(pv.Executable)
	})
	return
}

func (sum *summary) args(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		for _, item := range pv.Args {
			if match(item) {
				return true
			}
		}
		return false
	})
	return
}

func (sum *summary) user(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		return match(pv.Username)
	})
	return
}

func (sum *summary) cwd(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		return match(pv.Cwd)
	})
}

func (sum *summary) ppid(match func(string) bool) {
	if sum.init(); !sum.ok() {
		return
	}

	sum.view(func(pv *Process) bool {
		return match(auxlib.ToString(pv.Ppid))
	})
}

func (sum *summary) view(filter func(*Process) bool) {
	list := sum.List()
	n := len(list)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		pid := list[i]
		pv, err := Pid(pid)
		if err != nil {
			continue
		}

		if !filter(pv) {
			continue
		}
		sum.append(pv)
	}

}

func (sum *summary) search(cnd *cond.Cond) {
	list := sum.List()
	n := len(list)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		pid := list[i]
		pv, err := Pid(pid)
		if err != nil {
			continue
		}

		if cnd != nil && !cnd.Match(pv) {
			continue
		}
		sum.append(pv)
	}
}

func (sum *summary) checksum() {
	n := len(sum.Process)
	if n == 0 {
		return
	}

	distinct := make(map[string]string, n)

	for i := 0; i < n; i++ {
		p := sum.Process[i]
		if len(p.Executable) == 0 {
			continue
		}

		if hash, ok := distinct[p.Executable]; ok {
			p.Checksum = hash
			continue
		}

		p.Sha1()
		distinct[p.Executable] = p.Checksum
	}
}

func (sum *summary) GetByIndex(idx int) *Process {
	n := len(sum.Process)
	if n == 0 || idx > n || idx < 1 {
		return nil
	}

	return sum.Process[idx]
}

func (sum *summary) collect() {
	list := sum.List()
	n := len(list)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		pid := list[i]
		pv, err := Fast(pid)
		if err != nil {
			continue
		}
		sum.append(pv)
	}

}

func By(cnd *cond.Cond) *summary {
	sum := &summary{}
	sum.init()
	if !sum.ok() {
		xEnv.Infof("not found process summary %v", sum.Error)
		return sum
	}

	sum.search(cnd)
	return sum
}

func collect() *summary {
	sum := &summary{}
	sum.init()
	if !sum.ok() {
		xEnv.Infof("not found process summary %v", sum.Error)
		return sum
	}

	sum.collect()
	return sum
}

func NewSumL(L *lua.LState) *summary {
	return &summary{
		co:  L,
		vsh: vswitch.NewL(L),
	}
}
