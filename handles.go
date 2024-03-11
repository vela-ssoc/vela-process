package process

import (
	"encoding/json"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/vela-ssoc/vela-kit/fileutil"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"strings"
	"time"
)

type FileStats struct {
	Path  string    `json:"path"`
	Fd    uint64    `json:"fd"`
	CTime time.Time `json:"ctime"`
	ATime time.Time `json:"atime"`
	MTime time.Time `json:"Mtime"`
	Size  int64     `json:"size"`
}

func (f *FileStats) String() string                         { return lua.B2S(f.Byte()) }
func (f *FileStats) Type() lua.LValueType                   { return lua.LTObject }
func (f *FileStats) AssertFloat64() (float64, bool)         { return 0, false }
func (f *FileStats) AssertString() (string, bool)           { return "", false }
func (f *FileStats) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (f *FileStats) Peek() lua.LValue                       { return f }

func (f *FileStats) Byte() []byte {
	chunk, _ := json.Marshal(f)
	return chunk
}

func (f *FileStats) Lookup() error {
	if strings.HasPrefix(f.Path, "\\\\?\\") {
		f.Path = f.Path[4:]
	}
	atime, mtime, ctime, size, err := fileutil.State(f.Path)
	if err != nil {
		return err
	}

	f.CTime = ctime
	f.ATime = atime
	f.MTime = mtime
	f.Size = size
	return nil
}

func (f *FileStats) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "path":
		return lua.S2L(f.Path)
	case "fd":
		return lua.LInt64(int64(f.Fd))
	default:
		return lua.LNil
	}
}

type HandleSummary struct {
	Pid   int32                   `json:"pid"`
	Err   error                   `json:"err"`
	Files []process.OpenFilesStat `json:"files"`
}

func (hs *HandleSummary) String() string                         { return lua.B2S(hs.Byte()) }
func (hs *HandleSummary) Type() lua.LValueType                   { return lua.LTObject }
func (hs *HandleSummary) AssertFloat64() (float64, bool)         { return 0, false }
func (hs *HandleSummary) AssertString() (string, bool)           { return "", false }
func (hs *HandleSummary) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (hs *HandleSummary) Peek() lua.LValue                       { return hs }

func (hs *HandleSummary) Byte() []byte {
	chunk, _ := json.Marshal(hs)
	return chunk
}

func (hs *HandleSummary) pipeL(L *lua.LState) int {

	size := len(hs.Files)
	if size <= 0 {
		return 0
	}

	co := xEnv.Clone(L)
	defer xEnv.Free(co)

	chains := pipe.NewByLua(L)
	for i := 0; i < size; i++ {
		item := hs.Files[i]
		f := &FileStats{
			Path: item.Path,
			Fd:   item.Fd,
		}
		chains.Do(f, co, func(err error) {
			xEnv.Debugf("process handles pipe call fail %v", err)
		})
	}

	return 0

}

func (hs *HandleSummary) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "size":
		return lua.LInt(len(hs.Files))

	case "pipe":
		return lua.NewFunction(hs.pipeL)

	default:
		return lua.LNil
	}
}
