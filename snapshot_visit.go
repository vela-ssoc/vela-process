package process

import (
	"bytes"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/vela"
)

func (snt *snapshot) Create(bkt vela.Bucket) {
	var err error
	for pid, proc := range snt.current {
		sim := &simple{}
		if !proc.IsNull() {
			sim.with(proc)
			goto create
		}

		proc, err = sim.by(pid)
		if err != nil {
			continue
		}

		if snt.Ignore(proc) {
			continue
		}

	create:
		snt.report.OnCreate(proc)
		key := auxlib.ToString(pid)
		bkt.Store(key, sim, 0)
		snt.onCreate.Do(proc, snt.co, func(err error) {
			audit.Errorf("%s process snapshot create fail %v", snt.Name(), err).From(snt.co.CodeVM()).Put()
		})
	}

}

func (snt *snapshot) Delete(bkt vela.Bucket) {
	for pid, val := range snt.delete {
		bkt.Delete(pid)
		snt.onDelete.Do(val, snt.co, func(err error) {
			audit.Errorf("%s process snapshot delete fail %v", snt.Name(), err).From(snt.co.CodeVM()).Put()
		})
	}
}

func (snt *snapshot) Update(bkt vela.Bucket) {
	for pid, p := range snt.update {
		sim := &simple{}
		sim.with(p)
		bkt.Store(pid, sim, 0)
		snt.onUpdate.Do(p, snt.co, func(err error) {
			audit.Errorf("%s process snapshot update fail %v", snt.Name(), err).From(snt.co.CodeVM()).Put()
		})
	}
}

func (snt *snapshot) debug() {
	var buff bytes.Buffer
	bkt := xEnv.Bucket(snt.bkt...)
	bkt.Range(func(s string, i interface{}) {
		buff.WriteString(s)
		buff.WriteByte(':')
		buff.WriteString(auxlib.ToString(i))
		buff.WriteByte(',')
		buff.WriteByte('\n')
	})
	xEnv.Error(buff.String())
}
func (snt *snapshot) doReport() {
	if !snt.enable {
		return
	}

	snt.report.do()
}
