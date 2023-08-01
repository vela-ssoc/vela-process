package process

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/vela"
	"strings"
)

func view(ctx *fasthttp.RequestCtx) error {
	v := collect()
	//v.checksum()

	chunk, err := json.Marshal(v)
	if err != nil {
		return err
	}

	ctx.Write(chunk)
	return nil
}

func findId(ctx *fasthttp.RequestCtx) error {
	v := string(ctx.QueryArgs().Peek("pid"))
	if len(v) == 0 {
		return fmt.Errorf("got process pid emtpy")
	}

	pid, err := auxlib.ToInt32E(v)
	if err != nil {
		return err
	}

	p, err := Pid(pid)
	if err != nil {
		return err
	}
	p.Sha1()

	chunk, err := json.Marshal(p)
	if err != nil {
		return err
	}

	ctx.Write(chunk)
	return nil
}

func handle(ctx *fasthttp.RequestCtx) error {
	v := string(ctx.QueryArgs().Peek("pid"))
	if len(v) == 0 {
		return fmt.Errorf("got process pid emtpy")
	}

	pid, err := auxlib.ToInt32E(v)
	if err != nil {
		return err
	}

	p, err := Pid(pid)
	if err != nil {
		return err
	}

	fs, err := p.OpenFiles()
	if err != nil {
		return err
	}

	size := len(fs)
	data := make([]FileStats, 0, size)
	for i := 0; i < size; i++ {
		s := FileStats{
			Path: fs[i].Path,
			Fd:   fs[i].Fd,
		}

		if strings.HasPrefix(s.Path, "socket:") {
			continue
		}

		s.Lookup()
		data = append(data, s)
	}
	chunk, _ := json.Marshal(data)
	ctx.Write(chunk)
	return nil

}

func history(ctx *fasthttp.RequestCtx) error {
	bkt := xEnv.Bucket("VELA_FILE_HASH")
	ret := make(map[string]Checksum)
	bkt.ForEach(func(s string, chunk []byte) {
		var csm Checksum
		if len(chunk) == 0 {
			return
		}

		err := json.Unmarshal(chunk, &csm)
		if err != nil {
			xEnv.Errorf("%v file hash info decode fail %v", s, err)
			return
		}
		ret[s] = csm
	})

	body, err := json.Marshal(ret)
	if err != nil {
		return err
	}

	ctx.Write(body)
	return nil
}

func define(router vela.Router) {
	router.GET("/api/v1/arr/agent/process/list", xEnv.Then(view))
	router.GET("/api/v1/arr/agent/process/pid", xEnv.Then(findId))
	router.GET("/api/v1/arr/agent/process/handle", xEnv.Then(handle))
	router.GET("/api/v1/arr/agent/file/history", xEnv.Then(history))
}
