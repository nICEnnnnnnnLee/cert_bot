package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

func handleShell(name string, arg ...string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		flusher, _ := w.(http.Flusher)
		ctx := r.Context()

		var Fprintf = func(format string, a ...any) {
			fmt.Fprintf(w, format, a...)
			fmt.Fprintln(w)
			flusher.Flush()
		}
		var Fprintln = func(format string, a ...any) {
			fmt.Fprintf(w, format, a...)
			flusher.Flush()
		}
		c := exec.CommandContext(ctx, name, arg...)
		c.Env = os.Environ()
		stdout, err := c.StdoutPipe()
		if err != nil {
			Fprintf("%v", err)
			return
		}
		stderr, err := c.StderrPipe()
		if err != nil {
			Fprintf("%v", err)
			return
		}
		var wg sync.WaitGroup
		// 因为有2个任务, 一个需要读取stderr 另一个需要读取stdout
		wg.Add(2)
		go read(ctx, Fprintln, &wg, stderr)
		go read(ctx, Fprintln, &wg, stdout)
		// 这里一定要用start,而不是run 详情请看下面的图
		err = c.Start()
		if err != nil {
			Fprintf("%v", err)
			return
		}
		// 等待任务结束
		wg.Wait()
	}
}

func read(ctx context.Context, println func(format string, a ...any), wg *sync.WaitGroup, std io.ReadCloser) {
	reader := bufio.NewReader(std)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			println(readString)
		}
	}
}
