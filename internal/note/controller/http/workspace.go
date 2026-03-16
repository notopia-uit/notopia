package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func (h *StrictHandler) CreateWorkspace(
	ctx context.Context,
	request note.CreateWorkspaceRequestObject,
) (note.CreateWorkspaceResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) DeleteWorkspace(
	ctx context.Context,
	request note.DeleteWorkspaceRequestObject,
) (note.DeleteWorkspaceResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetWorkspace(
	ctx context.Context,
	request note.GetWorkspaceRequestObject,
) (note.GetWorkspaceResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetWorkspaceEvents(
	ctx context.Context,
	request note.GetWorkspaceEventsRequestObject,
) (note.GetWorkspaceEventsResponseObject, error) {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return nil, commonerror.NewInternal("failed to cast context to gin.Context", "", nil)
	}
	user, err := commonhttp.UserFromContextError(c)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	eventCh, err := h.workspaceEventPubSub.Subscribe(ctx, request.WorkspaceId, user.ID)
	if err != nil {
		return nil, commonerror.NewInternal("failed to subscribe to workspace events", "", err)
	}
	r, w := io.Pipe()
	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				slog.ErrorContext(c, "failed to close pipe writer in workspace events stream", slog.String("error", err.Error()))
			}
		}()
		for {
			select {
			case <-c.Done():
				return
			case <-ticker.C:
				if _, err := w.Write([]byte("heartbeat: keep-alive\n\n")); err != nil {
					slog.ErrorContext(c, "failed to write keep-alive comment in workspace events stream", slog.String("error", err.Error()))
					return
				}
				c.Writer.Flush()
				slog.DebugContext(c, "sent keep-alive comment in workspace events stream")
			case event, ok := <-eventCh:
				if !ok {
					slog.InfoContext(c, "workspace event channel closed")
					return
				}
				dto, ok := workspaceEventToDTO(event)
				if !ok {
					slog.WarnContext(c, "skipping unsupported workspace event type in workspace events stream", slog.String("event_type", event.EventType().String()))
					continue
				}
				eventBytes, err := json.Marshal(dto)
				if err != nil {
					slog.ErrorContext(c, "failed to marshal event to JSON", slog.String("error", err.Error()))
					continue
				}
				if _, err := w.Write([]byte("data: ")); err != nil {
					slog.ErrorContext(c, "failed to write event prefix in workspace events stream", slog.String("error", err.Error()))
					return
				}
				if _, err := w.Write(eventBytes); err != nil {
					slog.ErrorContext(c, "failed to write event data in workspace events stream", slog.String("error", err.Error()))
					return
				}
				if _, err := w.Write([]byte("\n\n")); err != nil {
					slog.ErrorContext(c, "failed to write event suffix in workspace events stream", slog.String("error", err.Error()))
					return
				}
				c.Writer.Flush()
			}
		}
	}()
	return note.GetWorkspaceEvents200TexteventStreamResponse{
		Body: r,
	}, nil
}

func (h *StrictHandler) CheckWorkspaceExists(
	ctx context.Context,
	request note.CheckWorkspaceExistsRequestObject,
) (note.CheckWorkspaceExistsResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetWorkspaceGraph(
	ctx context.Context,
	request note.GetWorkspaceGraphRequestObject,
) (note.GetWorkspaceGraphResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetWorkspaceMembers(
	ctx context.Context,
	request note.GetWorkspaceMembersRequestObject,
) (note.GetWorkspaceMembersResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) UpdateWorkspaceMembers(
	ctx context.Context,
	request note.UpdateWorkspaceMembersRequestObject,
) (note.UpdateWorkspaceMembersResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) MoveWorkspaceItems(
	ctx context.Context,
	request note.MoveWorkspaceItemsRequestObject,
) (note.MoveWorkspaceItemsResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetWorkspaceTree(
	ctx context.Context,
	request note.GetWorkspaceTreeRequestObject,
) (note.GetWorkspaceTreeResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) PublishWorkspace(
	ctx context.Context,
	request note.PublishWorkspaceRequestObject,
) (note.PublishWorkspaceResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) RenameWorkspace(
	ctx context.Context,
	request note.RenameWorkspaceRequestObject,
) (note.RenameWorkspaceResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) UnpublishWorkspace(
	ctx context.Context,
	request note.UnpublishWorkspaceRequestObject,
) (note.UnpublishWorkspaceResponseObject, error) {
	return nil, errors.New("not implemented")
}
