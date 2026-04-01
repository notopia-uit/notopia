package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type sseSender struct {
	ctx       context.Context
	writer    io.Writer
	flusher   http.Flusher
	mu        sync.Mutex
	closed    chan struct{}
	closeOnce sync.Once
}

func newSSESender(ctx context.Context, w io.Writer, flusher http.Flusher) *sseSender {
	return &sseSender{
		ctx:       ctx,
		writer:    w,
		flusher:   flusher,
		mu:        sync.Mutex{},
		closed:    make(chan struct{}),
		closeOnce: sync.Once{},
	}
}

func (s *sseSender) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.closed)
		if closer, ok := s.writer.(io.Closer); ok {
			err = closer.Close()
		}
	})
	return err
}

func (s *sseSender) Send(eventType string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString("event: ")
	buf.WriteString(eventType)
	buf.WriteString("\n")

	buf.WriteString("data: ")
	buf.Write(payload)
	buf.WriteString("\n\n")

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.writer.Write(buf.Bytes()); err != nil {
		return err
	}

	s.flusher.Flush()
	return nil
}

func (s *sseSender) StartHeartbeat(interval time.Duration, fn func() any) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.closed:
				return
			case <-ticker.C:
				_ = s.Send("heartbeat", fn())
			}
		}
	}()
}

func (s *sseSender) Stream(ch <-chan any, eventType string) {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.closed:
				return
			case evt, ok := <-ch:
				if !ok {
					return
				}
				_ = s.Send(eventType, evt)
			}
		}
	}()
}
