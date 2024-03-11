package process

import (
	"bytes"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/pipe"
	"github.com/vela-ssoc/vela-kit/strutil"
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

func (sa *snapshot) LookupAll(pex *ProcEx) (*Process, error) {

	if cwd, err := pex.ps.Cwd(); err == nil {
		pex.value.Cwd = cwd
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
	return pex.value, nil
}

func (sa *snapshot) Create(bkt vela.Bucket) {
	for pid, _ := range sa.current {
		pex, err := sa.simple(pid)
		if err != nil {
			//xEnv.Errorf("not found pid=%d process on create fail %v", pid, err)
			continue
		}

		proc, err := sa.LookupAll(pex)
		if err != nil {
			xEnv.Errorf("pid=%d process lookup all fail %v", pid, err)
		}

		//if sa.Ignore(proc) {
		//	continue
		//}
		//sa.vsh.Do(proc)
		//key := strutil.String(pid)
		//bkt.Store(key, proc, 0)

		sa.report.OnCreate(proc)
		//sa.onCreate.Do(proc, sa.co, func(err error) {
		//	audit.Errorf("%s process snapshot create fail %v", sa.Name(), err).From(sa.co.CodeVM()).Put()
		//})
	}

}

func (sa *snapshot) Delete(bkt vela.Bucket) {
	for pid, val := range sa.delete {
		if e := bkt.Delete(pid); e != nil {
			xEnv.Errorf("delete process pid:%s val: %v fail %v", pid, val, e)
		}
		//sa.onDelete.Do(val, sa.co, func(err error) {
		//	audit.Errorf("%s process snapshot delete fail %v", sa.Name(), err).From(sa.co.CodeVM()).Put()
		//})
	}
}

func (sa *snapshot) Update(bkt vela.Bucket) {
	for pid, pex := range sa.update {
		proc, err := sa.LookupAll(pex)
		if err != nil {
			xEnv.Errorf("not found process info %v", err)
		}

		key := strutil.String(pid)
		bkt.Store(key, proc, 0)
		sa.vsh.Do(proc)
		sa.report.OnUpdate(pex.value)
		//sa.onUpdate.Do(proc, sa.co, func(err error) {
		//	audit.Errorf("%s process snapshot update fail %v", sa.Name(), err).From(sa.co.CodeVM()).Put()
		//})
	}
}

func (sa *snapshot) debug() {
	var buff bytes.Buffer
	bkt := xEnv.Shm(V_PROC_SHM)
	bkt.Range(func(s string, i interface{}) {
		buff.WriteString(s)
		buff.WriteByte(':')
		buff.WriteString(auxlib.ToString(i))
		buff.WriteByte(',')
		buff.WriteByte('\n')
	})
	xEnv.Error(buff.String())
}

func (sa *snapshot) tree(ppid int32, treeEx *ProcTree) {
	if ppid == 0 {
		return
	}

	pex, ok := sa.factory[ppid]
	if !ok {
		return
	}

	treeEx.Add(pex.value.Pid, pex.value.Name)

	sa.tree(pex.value.Ppid, treeEx)
}

func (sa *snapshot) Tree(proc *Process) ProcTree {
	treeEx := ProcTree{
		Pids: []int32{proc.Pid},
		Tree: []string{proc.Name},
	}

	sa.tree(proc.Ppid, &treeEx)
	return treeEx
}

func (sa *snapshot) doReport(bkt vela.Bucket) {

	fnc := func(proc *Process, chain *pipe.Chains, flag bool) {
		proc.PidTree = sa.Tree(proc)
		key := strutil.String(proc.Pid)
		if flag {
			_ = bkt.Store(key, proc, 0)
		}

		if sa.Ignore(proc) {
			return
		}

		sa.vsh.Do(proc)
		chain.Do(proc, sa.co, func(err error) {
			xEnv.Errorf("%s snapshot call fail %v", sa.Name(), err)
		})
	}

	n := len(sa.report.Creates)
	for i := 0; i < n; i++ {
		proc := sa.report.Creates[i]
		fnc(proc, sa.onCreate, true)
	}

	n = len(sa.report.Updates)
	for i := 0; i < n; i++ {
		proc := sa.report.Updates[i]
		fnc(proc, sa.onUpdate, true)
	}

	n = len(sa.report.Deletes)
	for i := 0; i < n; i++ {
		proc := sa.report.Deletes[i]
		fnc(proc, sa.onDelete, false)
	}

	if !sa.enable {
		return
	}

	sa.report.do()
}
