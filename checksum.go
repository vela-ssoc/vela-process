package process

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vela-ssoc/vela-kit/fileutil"
	"io"
	"os"
)

type Checksum struct {
	ATime int64  `json:"a_time"`
	MTime int64  `json:"m_time"`
	CTime int64  `json:"c_time"`
	Sha1  string `json:"sha1"`
	Md5   string `json:"md5"`
	Size  int64  `json:"size"`
}

func (csm *Checksum) Decode(v []byte) error {
	if len(v) == 0 {
		return fmt.Errorf("emtpy value")
	}
	return json.Unmarshal(v, csm)
}

func (csm *Checksum) Encode() ([]byte, error) {
	return json.Marshal(csm)
}

func hash(exe string) (*Checksum, error) {
	csm := &Checksum{}
	bkt := xEnv.Bucket("VELA_FILE_HASH")
	body, _ := bkt.Value(exe)

	if len(body) > 0 && csm.Decode(body) == nil {
		//xEnv.Infof("decode body fail %v", e)
	}

	info, err := os.Stat(exe)
	if err != nil {
		return csm, err
	}

	if info.Size() == csm.Size && info.ModTime().Unix() == csm.MTime {
		return csm, nil
	}

	fd, err := os.Open(exe)
	if err != nil {
		return csm, err
	}
	defer fd.Close()

	csm.MTime = info.ModTime().Unix()
	atime, mtime, ctime, size, err := fileutil.StateByInfo(info)
	if err != nil {
		return csm, err
	}

	sh1 := sha1.New()
	m5 := md5.New()

	w := io.MultiWriter(m5, sh1)
	_, err = io.Copy(w, fd)
	if err != nil {
		return csm, err
	}

	csm.Md5 = hex.EncodeToString(m5.Sum(nil))
	csm.Sha1 = hex.EncodeToString(sh1.Sum(nil))
	csm.Size = size
	csm.MTime = mtime.Unix()
	csm.ATime = atime.Unix()
	csm.CTime = ctime.Unix()

	body, err = csm.Encode()
	if err == nil {
		bkt.Store(exe, body, 0)
	}

	return csm, nil
}
