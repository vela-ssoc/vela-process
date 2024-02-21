package process

import (
	"bytes"
	"encoding/gob"
)

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

	var p Process
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&p)
	return &p, err
}
