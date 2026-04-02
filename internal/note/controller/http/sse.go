package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type workspaceEventSSESender struct {
	ctx       context.Context
	eventCh   <-chan app.WorkspaceEvent
	writer    io.Writer
	flusher   http.Flusher
	mu        sync.Mutex
	closeOnce sync.Once
}

func newWorkspaceEventSSESender(
	ctx context.Context,
	eventCh <-chan app.WorkspaceEvent,
	w io.Writer,
	flusher http.Flusher,
) *workspaceEventSSESender {
	return &workspaceEventSSESender{
		ctx:       ctx,
		eventCh:   eventCh,
		writer:    w,
		flusher:   flusher,
		mu:        sync.Mutex{},
		closeOnce: sync.Once{},
	}
}

func (s *workspaceEventSSESender) send(event app.WorkspaceEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString("id: ")
	buf.WriteString(event.GetID().String())
	buf.WriteString("\n")

	buf.WriteString("event: ")
	buf.WriteString(event.GetEvent())
	buf.WriteString("\n")

	buf.WriteString("data: ")
	buf.Write(payload)
	buf.WriteString("\n\n")

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.writer.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write SSE event: %v", err)
	}

	s.flusher.Flush()
	return nil
}

func (s *workspaceEventSSESender) sendHeartBeat() error {
	event := note.HeartBeatWorkspaceEvent{
		Event:     note.HeartBeatWorkspaceEventEventHeartBeatWorkspaceEvent,
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat event: %v", err)
	}
	heartBeatPayload := fmt.Sprintf("event: %s\ndata: %s\n\n", note.HeartBeatWorkspaceEventEventHeartBeatWorkspaceEvent, payload)
	s.mu.Lock()
	if _, err := s.writer.Write([]byte(heartBeatPayload)); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to write heartbeat SSE event: %v", err)
	}
	s.flusher.Flush()
	s.mu.Unlock()
	return nil
}

func (s *workspaceEventSSESender) Stream() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				if err := s.sendHeartBeat(); err != nil {
					slog.Warn("failed to send heartbeat", "error", err)
				}
				slog.Debug("sent heartbeat")
				continue
			case evt, ok := <-s.eventCh:
				if !ok {
					return
				}
				_ = s.send(evt)
			}
		}
	}()
}
