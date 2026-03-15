package http

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/api/note"
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
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}
	eventCh := make(chan []byte)
	h.workspaceEventHub.Subscribe(request.WorkspaceId, user.ID, eventCh)
	r, w := io.Pipe()
	go func() {
		defer h.workspaceEventHub.Unsubscribe(request.WorkspaceId, user.ID)
		defer func() {
			if err := w.Close(); err != nil {
				slog.ErrorContext(ctx, "failed to close pipe writer in workspace events stream", slog.String("error", err.Error()))
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-eventCh:
				if !ok {
					slog.InfoContext(ctx, "workspace event channel closed")
					return
				}
				if _, err := w.Write([]byte("data: ")); err != nil {
					slog.ErrorContext(ctx, "failed to write event prefix in workspace events stream", slog.String("error", err.Error()))
					return
				}
				if _, err := w.Write(event); err != nil {
					slog.ErrorContext(ctx, "failed to write event data in workspace events stream", slog.String("error", err.Error()))
					return
				}
				if _, err := w.Write([]byte("\n\n")); err != nil {
					slog.ErrorContext(ctx, "failed to write event suffix in workspace events stream", slog.String("error", err.Error()))
					return
				}
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
