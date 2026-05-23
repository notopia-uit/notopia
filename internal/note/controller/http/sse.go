package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type WorkspaceEventStreamResponse struct {
	ctx     context.Context
	eventCh <-chan app.WorkspaceEvent
}

func NewWorkspaceEventStreamResponse(
	ctx context.Context,
	eventCh <-chan app.WorkspaceEvent,
) *WorkspaceEventStreamResponse {
	return &WorkspaceEventStreamResponse{
		ctx:     ctx,
		eventCh: eventCh,
	}
}

var _ note.GetWorkspaceEventsResponseObject = (*WorkspaceEventStreamResponse)(nil)

func (r *WorkspaceEventStreamResponse) VisitGetWorkspaceEventsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flushing")
	}
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return nil
		case <-ticker.C:
			if err := writeHeartBeat(w); err != nil {
				slog.Warn("failed to send heartbeat", slog.Any("error", err))
				return nil
			}
		case evt, ok := <-r.eventCh:
			if !ok {
				return nil
			}
			if err := writeWorkspaceEvent(w, evt); err != nil {
				slog.Error("failed to send event", slog.Any("error", err))
				return nil
			}
		}

		flusher.Flush()
	}
}

func writeHeartBeat(w http.ResponseWriter) error {
	event := note.HeartBeatWorkspaceEvent{
		Event:     note.HeartBeatWorkspaceEventEventHeartBeatWorkspaceEvent,
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", note.HeartBeatWorkspaceEventEventHeartBeatWorkspaceEvent, payload)
	return err
}

func writeWorkspaceEvent(w http.ResponseWriter, event app.WorkspaceEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	_, err = fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", event.GetID().String(), event.GetEvent(), payload)
	return err
}
