package commonhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ref: https://github.com/notopia-uit/notopia/issues/48
type SSEWrapper[T any] struct {
	Events <-chan T
	Ctx    context.Context
}

var (
	_ io.Reader   = (*SSEWrapper[any])(nil)
	_ io.WriterTo = (*SSEWrapper[any])(nil)
)

func (s SSEWrapper[T]) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (s SSEWrapper[T]) WriteTo(w io.Writer) (int64, error) {
	f, ok := w.(http.Flusher)
	if !ok {
		return 0, errors.New("streaming unsupported")
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")
		rw.Header().Set("Transfer-Encoding", "chunked")
	}

	for {
		select {
		case <-s.Ctx.Done():
			return 0, s.Ctx.Err()
		case event, ok := <-s.Events:
			if !ok {
				return 0, nil
			}

			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
			if err != nil {
				return 0, err
			}

			f.Flush()
		}
	}
}
