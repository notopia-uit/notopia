package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) GetWorkspaceEvents(
	ctx context.Context,
	request note.GetWorkspaceEventsRequestObject,
) (note.GetWorkspaceEventsResponseObject, error) {
	// eventChan := make(chan any)
	// go s.streamFromRedis(request.WorkspaceId, eventChan, ctx)
	return &note.GetWorkspaceEvents200TexteventStreamResponse{}, nil
}
