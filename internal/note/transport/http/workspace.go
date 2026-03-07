package http

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/pkg/api/note"
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
	// eventChan := make(chan any)
	// go s.streamFromRedis(request.WorkspaceId, eventChan, ctx)
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetWorkspaceGraph(
	ctx context.Context,
	request note.GetWorkspaceGraphRequestObject,
) (note.GetWorkspaceGraphResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) MoveWorkspaceItems(
	ctx context.Context,
	request note.MoveWorkspaceItemsRequestObject,
) (note.MoveWorkspaceItemsResponseObject, error) {
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

func (h *StrictHandler) UpdateWorkspaceCollaborators(
	ctx context.Context,
	request note.UpdateWorkspaceCollaboratorsRequestObject,
) (note.UpdateWorkspaceCollaboratorsResponseObject, error) {
	return nil, errors.New("not implemented")
}
