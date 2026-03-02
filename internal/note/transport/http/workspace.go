package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) CreateWorkspace(
	ctx context.Context,
	request note.CreateWorkspaceRequestObject,
) (note.CreateWorkspaceResponseObject, error) {
	return &note.CreateWorkspace201Response{}, nil
}

func (h *StrictHandler) DeleteWorkspace(
	ctx context.Context,
	request note.DeleteWorkspaceRequestObject,
) (note.DeleteWorkspaceResponseObject, error) {
	return &note.DeleteWorkspace204Response{}, nil
}

func (h *StrictHandler) GetWorkspace(
	ctx context.Context,
	request note.GetWorkspaceRequestObject,
) (note.GetWorkspaceResponseObject, error) {
	return &note.GetWorkspace200JSONResponse{}, nil
}

func (h *StrictHandler) PublishWorkspace(
	ctx context.Context,
	request note.PublishWorkspaceRequestObject,
) (note.PublishWorkspaceResponseObject, error) {
	return &note.PublishWorkspace204Response{}, nil
}

func (h *StrictHandler) RenameWorkspace(
	ctx context.Context,
	request note.RenameWorkspaceRequestObject,
) (note.RenameWorkspaceResponseObject, error) {
	return &note.RenameWorkspace204Response{}, nil
}

func (h *StrictHandler) UnpublishWorkspace(
	ctx context.Context,
	request note.UnpublishWorkspaceRequestObject,
) (note.UnpublishWorkspaceResponseObject, error) {
	return &note.UnpublishWorkspace204Response{}, nil
}

func (h *StrictHandler) GetWorkspaceEvents(
	ctx context.Context,
	request note.GetWorkspaceEventsRequestObject,
) (note.GetWorkspaceEventsResponseObject, error) {
	// eventChan := make(chan any)
	// go s.streamFromRedis(request.WorkspaceId, eventChan, ctx)
	return &note.GetWorkspaceEvents200TexteventStreamResponse{}, nil
}
