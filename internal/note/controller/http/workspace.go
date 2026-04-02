package http

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) CreateWorkspace(
	ctx context.Context,
	request note.CreateWorkspaceRequestObject,
) (note.CreateWorkspaceResponseObject, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, errs.NewInternalGenerateID(err)
	}
	cmd := &app.CreateWorkspace{
		ID:   id,
		Name: request.Body.Name,
		Slug: request.Body.Slug,
	}
	err = h.App.CommandHandlers.CreateWorkspaceHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateWorkspace201Response{
		Headers: note.CreateWorkspace201ResponseHeaders{
			ContentLocation: h.BaseURL.JoinPath("workspaces", id.String()).String(),
		},
	}, nil
}

func (h *StrictHandler) DeleteWorkspace(
	ctx context.Context,
	request note.DeleteWorkspaceRequestObject,
) (note.DeleteWorkspaceResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.NewUnauthorized()
	}

	cmd := &app.DeleteWorkspace{
		ID:     request.WorkspaceId,
		UserID: user.ID,
	}
	err := h.App.CommandHandlers.DeleteWorkspaceHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.DeleteWorkspace204Response{}, nil
}

func (h *StrictHandler) GetWorkspace(
	ctx context.Context,
	request note.GetWorkspaceRequestObject,
) (note.GetWorkspaceResponseObject, error) {
	query := &app.GetWorkspaceBySlug{
		Slug: request.WorkspaceSlug,
	}
	result, err := h.App.QueryHandlers.GetWorkspaceHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toWorkspace(*result)
	return note.GetWorkspace200JSONResponse(dto), nil
}

func (h *StrictHandler) CheckWorkspaceSlugExists(
	ctx context.Context,
	request note.CheckWorkspaceSlugExistsRequestObject,
) (note.CheckWorkspaceSlugExistsResponseObject, error) {
	query := &app.CheckWorkspaceSlugExists{
		Slug: request.WorkspaceSlug,
	}
	result, err := h.App.QueryHandlers.CheckWorkspaceSlugExistsHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Exists {
		return note.CheckWorkspaceSlugExists409Response{}, nil
	}
	return note.CheckWorkspaceSlugExists200Response{}, nil
}

func (h *StrictHandler) GetWorkspaceEvents(
	ctx context.Context,
	request note.GetWorkspaceEventsRequestObject,
) (note.GetWorkspaceEventsResponseObject, error) {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return nil, errs.NewInternal("failed to cast context to gin.Context", nil)
	}
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.NewUnauthorized()
	}

	eventCh, err := h.WorkspaceEventPubSub.Subscribe(ctx, request.WorkspaceId, user.ID)
	if err != nil {
		return nil, errs.NewInternal("failed to subscribe to workspace events", err)
	}
	r, w := io.Pipe()
	sender := newWorworkspaceEventSSESender(ctx, eventCh, w, c.Writer)
	sender.Stream()

	//exhaustruct:ignore
	return note.GetWorkspaceEvents200TexteventStreamResponse{
		Body: r,
	}, nil
}

func (h *StrictHandler) GetWorkspaceGraph(
	ctx context.Context,
	request note.GetWorkspaceGraphRequestObject,
) (note.GetWorkspaceGraphResponseObject, error) {
	ignoreOrphans := false
	if request.Params.IncludeOrphans != nil {
		ignoreOrphans = !*request.Params.IncludeOrphans
	}
	query := &app.GetWorkspaceGraph{
		ID:            request.WorkspaceId,
		IgnoreOrphans: ignoreOrphans,
	}
	result, err := h.App.QueryHandlers.GetWorkspaceGraphHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toGraph(result)
	return note.GetWorkspaceGraph200JSONResponse(dto), nil
}

func (h *StrictHandler) GetWorkspaceMembers(
	ctx context.Context,
	request note.GetWorkspaceMembersRequestObject,
) (note.GetWorkspaceMembersResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) UpdateWorkspaceMembers(
	ctx context.Context,
	request note.UpdateWorkspaceMembersRequestObject,
) (note.UpdateWorkspaceMembersResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) MoveWorkspaceItems(
	ctx context.Context,
	request note.MoveWorkspaceItemsRequestObject,
) (note.MoveWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.NewUnauthorized()
	}

	var noteIDs []uuid.UUID
	if request.Body.NoteIds != nil {
		noteIDs = *request.Body.NoteIds
	}

	var folderIDs []uuid.UUID
	if request.Body.FolderIds != nil {
		folderIDs = *request.Body.FolderIds
	}

	var destFolderID uuid.UUID
	if request.Body.DestinationFolderId != nil {
		destFolderID = *request.Body.DestinationFolderId
	}

	cmd := &app.MoveWorkspaceItems{
		UserID:              user.ID,
		WorkspaceID:         request.WorkspaceId,
		NoteIDs:             noteIDs,
		FolderIDs:           folderIDs,
		DestinationFolderID: destFolderID,
	}
	err := h.App.CommandHandlers.MoveWorkspaceItemsHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.MoveWorkspaceItems204Response{}, nil
}

func (h *StrictHandler) PublishWorkspace(
	ctx context.Context,
	request note.PublishWorkspaceRequestObject,
) (note.PublishWorkspaceResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) RenameWorkspace(
	ctx context.Context,
	request note.RenameWorkspaceRequestObject,
) (note.RenameWorkspaceResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.NewUnauthorized()
	}

	cmd := &app.RenameWorkspace{
		ID:     request.WorkspaceId,
		Name:   request.Body.Name,
		UserID: user.ID,
	}
	err := h.App.CommandHandlers.RenameWorkspaceHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.RenameWorkspace204Response{}, nil
}

func (h *StrictHandler) RestoreTrashedWorkspaceItems(
	ctx context.Context,
	request note.RestoreTrashedWorkspaceItemsRequestObject,
) (note.RestoreTrashedWorkspaceItemsResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) ShowTrash(
	ctx context.Context,
	request note.ShowTrashRequestObject,
) (note.ShowTrashResponseObject, error) {
	query := &app.ShowTrash{
		WorkspaceID: request.WorkspaceId,
	}
	result, err := h.App.QueryHandlers.ShowTrashHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toShowTrash(result)
	return note.ShowTrash200JSONResponse(dto), nil
}

func (h *StrictHandler) TrashWorkspaceItems(
	ctx context.Context,
	request note.TrashWorkspaceItemsRequestObject,
) (note.TrashWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.NewUnauthorized()
	}

	cmd := &app.TrashWorkspaceItems{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
		NoteIDs:     *request.Body.NoteIds,
		FolderIDs:   *request.Body.FolderIds,
	}
	err := h.App.CommandHandlers.TrashWorkspaceItemsHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.TrashWorkspaceItems204Response{}, nil
}

func (h *StrictHandler) GetWorkspaceTree(
	ctx context.Context,
	request note.GetWorkspaceTreeRequestObject,
) (note.GetWorkspaceTreeResponseObject, error) {
	var depth *uint
	if request.Params.Depth != nil && *request.Params.Depth > 0 {
		depth = new(uint(*request.Params.Depth))
	}

	query := &app.GetWorkspaceTree{
		WorkspaceID:    request.WorkspaceId,
		RootFolderID:   request.Params.RootFolderId,
		IncludeTrashed: request.Params.IncludeTrashed != nil && *request.Params.IncludeTrashed,
		Depth:          depth,
	}

	result, err := h.App.QueryHandlers.GetWorkspaceTreeHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toWorkspaceTreeFolder(result)
	return note.GetWorkspaceTree200JSONResponse(dto), nil
}

func (h *StrictHandler) UnpublishWorkspace(
	ctx context.Context,
	request note.UnpublishWorkspaceRequestObject,
) (note.UnpublishWorkspaceResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) PermanentlyDeleteWorkspaceItems(
	ctx context.Context,
	request note.PermanentlyDeleteWorkspaceItemsRequestObject,
) (note.PermanentlyDeleteWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.NewUnauthorized()
	}

	var noteIDs []uuid.UUID
	if request.Body.NoteIds != nil {
		noteIDs = *request.Body.NoteIds
	}

	var folderIDs []uuid.UUID
	if request.Body.FolderIds != nil {
		folderIDs = *request.Body.FolderIds
	}

	cmd := &app.PermanentlyDeleteWorkspaceItems{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
		NoteIDs:     noteIDs,
		FolderIDs:   folderIDs,
	}
	err := h.App.CommandHandlers.PermanentlyDeleteWorkspaceItemsHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.PermanentlyDeleteWorkspaceItems204Response{}, nil
}
