package process

import (
	"bytes"
	"encoding/gob"
	"strings"
)

type null struct{}

var NULL = null{}

type simple Process

/*
type simple struct {
	Name       string   `json:"name"`
	State      string   `json:"state"`
	Pid        int      `json:"pid"`
	PPid       int      `json:"ppid"`
	PGid       uint32   `json:"pgid"`
	Cmdline    string   `json:"cmdline"`
	Username   string   `json:"username"`
	Cwd        string   `json:"cwd"`
	Executable string   `json:"executable"` // linux
	Args       []string `json:"args"`
}

func (s *simple) with(p *Process) {
	s.Name = p.Name
	s.State = p.State
	s.Pid = p.Pid
	s.PPid = p.Ppid
	s.PGid = p.Pgid
	s.Cmdline = p.Cmdline
	s.Username = p.Username
	s.Cwd = p.Cwd
	s.Executable = p.Executable
	s.Args = p.Args
}
*/

func (s *simple) with(p *Process) {
	*s = simple(*p)
}

func (s *simple) binary() string {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		return ""
		xEnv.Errorf("pid:%d name:%s gob encode fail %v", s.Pid, s.Name, err)
	}

	return buf.String()
}

func (s *simple) ArgsToString() string {
	return strings.Join(s.Args, " ")
}

func (s *simple) Equal(old *simple) bool {
	switch {
	case s.Name != old.Name:
		return false
	case s.Ppid != old.Ppid:
		return false
	case s.Cmdline != old.Cmdline:
		return false
	case s.Username != old.Username:
		return false
	case s.Cwd != old.Cwd:
		return false
	case s.Executable != old.Executable:
		return false
	case s.Mtime != old.Mtime:
		return false

	case s.Uptime != old.Uptime:
		return false
	}
	return true
}

func (s *simple) exe() string {
	if s.Executable != "" {
		return s.Executable
	}

	return ""
}

func encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
		//xEnv.Errorf("pid:%d name:%s gob encode fail %v", s.Pid, s.Name, err)
		//return ""
	}
	return buf.Bytes(), nil
}

func decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var s simple
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&s)
	return &s, err
}
