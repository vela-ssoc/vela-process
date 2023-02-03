package process

import (
	"github.com/vela-ssoc/vela-kit/opcode"
)

type report struct {
	Deletes []int      `json:"deletes"`
	Updates []*Process `json:"updates"`
	Creates []*Process `json:"creates"`
}

func (r *report) OnCreate(p *Process) {
	p.Sha1()
	p.Snap = "create"
	r.Creates = append(r.Creates, p)
}

func (r *report) OnUpdate(p *Process) {
	p.Sha1()
	p.Snap = "update"
	r.Updates = append(r.Updates, p)
}

func (r *report) OnDelete(p int) {
	r.Deletes = append(r.Deletes, p)
}

func (r *report) Len() int {
	return len(r.Updates) + len(r.Deletes) + len(r.Creates)
}

func (r *report) do() {
	if r.Len() == 0 {
		return
	}
	op := opcode.OpProcessDiff
	err := xEnv.TnlSend(op, r)
	if err != nil {
		xEnv.Errorf("tunnel send push opcode:%d fail %v", op, err)
	}
}
