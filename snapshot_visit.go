package process

import (
	"bytes"
	"fmt"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/vela"
	"strings"
	"time"
)

func (pex *ProcEx) Name() string {
	if pex.value.Name != "" {
		return pex.value.Name
	}

	if exe, err := pex.ps.Name(); err == nil {
		pex.value.Name = exe
		return exe
	}

	pex.value.Name = ""
	return ""

}

func (pex *ProcEx) Exe() string {
	if pex.value.Executable != "" {
		return pex.value.Executable
	}

	if exe, err := pex.ps.Exe(); err == nil {
		pex.value.Executable = exe
		return exe
	}

	pex.value.Executable = "N"
	return "N"
}

func (pex *ProcEx) Cmdline() string {
	if pex.value.Cmdline != "" {
		return pex.value.Cmdline
	}

	if cmdline, err := pex.ps.Cmdline(); err == nil {
		pex.value.Cmdline = cmdline
		pex.value.Args = strings.Split(cmdline, " ")
		return cmdline
	}

	return ""
}

func (sa *snapshot) Factory(pid int32, pex *ProcEx) {
	if pex.value.Ppid == 0 &&
		pex.value.Name == "kthreadd" &&
		pex.value.Username == "root" &&
		pex.value.Executable == "" {
		sa.system[pid] = "[kthreadd]"
	}
	sa.factory[pid] = pex
}

func (sa *snapshot) LookupParent(pex *ProcEx) (*ProcEx, error) {

	var err error

	ppex, ok := sa.factory[pex.value.Ppid]
	if !ok {
		ppex, err = sa.simple(pex.value.Ppid)
		if err != nil {
			//xEnv.Errorf("not found pid=%d parent process fail %v", pex.value.Pid, err)
			return nil, err
		}
	}

	pex.value.ParentName = ppex.Name()
	pex.value.ParentExecutable = ppex.Exe()
	pex.value.ParentCmdline = ppex.Cmdline()
	pex.value.ParentUsername = ppex.value.Username
	return ppex, nil
}

func (sa *snapshot) LookupPidTree(pex *ProcEx, tree []string) []string {
	if pex == nil || pex.value.Ppid == 0 {
		return tree
	}

	var err error
	var ppex *ProcEx

	ppex, err = sa.LookupParent(pex)
	if err != nil {
		return tree
	}

	tree = append([]string{fmt.Sprintf("%d.%s", pex.value.Ppid, pex.value.ParentName)}, tree...)
	tree = sa.LookupPidTree(ppex, tree)

	return tree
}

func (sa *snapshot) LookupAll(pex *ProcEx) (*Process, error) {

	if cwd, err := pex.ps.Cwd(); err == nil {
		pex.value.Cwd = cwd
	}

	if status, err := pex.ps.Status(); err == nil {
		pex.value.State = status
	}

	if exe, err := pex.ps.Exe(); err == nil {
		pex.value.Executable = exe
	}

	if cmdline, e := pex.ps.Cmdline(); e == nil {
		pex.value.Cmdline = cmdline
		pex.value.Args = strings.Split(cmdline, " ")
	}

	if pex.value.Executable != "" {
		csm, err := hash(pex.value.Executable)
		if err == nil {
			pex.value.Checksum = csm.Sha1
			pex.value.Md5 = csm.Md5
			pex.value.Mtime = time.Unix(csm.MTime, 0)
			pex.value.Ctime = time.Unix(csm.CTime, 0)
		} else {
			xEnv.Errorf("executable %s hash compute fail %v", pex.value.Executable, err)
		}
	}

	if pex.value.Ppid == pex.value.Pid {
		return pex.value, nil
	}

	pex.value.PidTree = sa.LookupPidTree(pex, []string{fmt.Sprintf("%d.%s", pex.pid, pex.Name())})

	return pex.value, nil
}

func (sa *snapshot) Create(bkt vela.Bucket) {
	for pid, _ := range sa.current {
		key := auxlib.ToString(pid)
		pex, err := sa.simple(pid)
		if err != nil {
			//xEnv.Errorf("not found pid=%d process on create fail %v", pid, err)
			continue
		}

		proc, err := sa.LookupAll(pex)
		if err != nil {
			xEnv.Errorf("pid=%d process lookup all fail %v", pid, err)
		}

		if sa.Ignore(proc) {
			continue
		}
		sa.vsh.Do(proc)
		sa.report.OnCreate(proc)
		bkt.Store(key, proc, 0)
		sa.onCreate.Do(proc, sa.co, func(err error) {
			audit.Errorf("%s process snapshot create fail %v", sa.Name(), err).From(sa.co.CodeVM()).Put()
		})
	}

}

func (sa *snapshot) Delete(bkt vela.Bucket) {
	for pid, val := range sa.delete {
		if e := bkt.Delete(pid); e != nil {
			xEnv.Errorf("delete process pid:%s val: %v fail %v", pid, val, e)
		}
		sa.onDelete.Do(val, sa.co, func(err error) {
			audit.Errorf("%s process snapshot delete fail %v", sa.Name(), err).From(sa.co.CodeVM()).Put()
		})
	}
}

func (sa *snapshot) Update(bkt vela.Bucket) {
	for pid, pex := range sa.update {
		proc, err := sa.LookupAll(pex)
		if err != nil {
			xEnv.Errorf("not found process info %v", err)
		}

		key := auxlib.ToString(pid)
		bkt.Store(key, proc, 0)
		sa.vsh.Do(proc)
		sa.report.OnUpdate(pex.value)
		sa.onUpdate.Do(proc, sa.co, func(err error) {
			audit.Errorf("%s process snapshot update fail %v", sa.Name(), err).From(sa.co.CodeVM()).Put()
		})
	}
}

func (sa *snapshot) debug() {
	var buff bytes.Buffer
	bkt := xEnv.Bucket(sa.bkt...)
	bkt.Range(func(s string, i interface{}) {
		buff.WriteString(s)
		buff.WriteByte(':')
		buff.WriteString(auxlib.ToString(i))
		buff.WriteByte(',')
		buff.WriteByte('\n')
	})
	xEnv.Error(buff.String())
}

func (sa *snapshot) doReport() {
	if !sa.enable {
		return
	}

	sa.report.do()
}
