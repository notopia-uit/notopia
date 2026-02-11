package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/note"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func (h *StrictHandler) GetWorkspaceEvents(
	ctx context.Context,
	request note.GetWorkspaceEventsRequestObject,
) (note.GetWorkspaceEventsResponseObject, error) {
	// eventChan := make(chan any)
	// go s.streamFromRedis(request.WorkspaceId, eventChan, ctx)
	return &note.GetWorkspaceEvents200TexteventStreamResponse{
		GetWorkspaceEventsResponseTexteventStreamResponse: note.GetWorkspaceEventsResponseTexteventStreamResponse{
			Body: commonhttp.SSEWrapper[any]{
				Events: nil,
				Ctx:    ctx,
			},
		},
	}, nil
}
