package process

type report struct {
	Deletes []int32    `json:"deletes"`
	Updates []*Process `json:"updates"`
	Creates []*Process `json:"creates"`
}

func (r *report) OnCreate(p *Process) {
	p.Snap = "create"
	r.Creates = append(r.Creates, p)
}

func (r *report) OnUpdate(p *Process) {
	p.Snap = "update"
	r.Updates = append(r.Updates, p)
}

func (r *report) OnDelete(p int32) {
	r.Deletes = append(r.Deletes, p)
}

func (r *report) Len() int {
	return len(r.Updates) + len(r.Deletes) + len(r.Creates)
}

func (r *report) do() {
	if r.Len() == 0 {
		return
	}
	err := xEnv.Push("/api/v1/broker/collect/agent/process/diff", r)
	if err != nil {
		xEnv.Errorf("tunnel send push diff fail %v", err)
	}
}
