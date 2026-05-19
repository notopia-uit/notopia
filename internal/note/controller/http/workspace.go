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
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, errs.NewInternalGenerateID(err)
	}
	cmd := &app.CreateWorkspace{
		ID:      id,
		Name:    request.Body.Name,
		Slug:    request.Body.Slug,
		OwnerID: user.ID,
	}
	err = h.App.Cmds.CreateWorkspace.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateWorkspace201JSONResponse{
		Headers: note.CreateWorkspace201ResponseHeaders{
			ContentLocation: h.BaseURL.JoinPath("note", "workspaces", id.String()).String(),
		},
		Body: note.CreateWorkspaceResponse{
			Id: &id,
		},
	}, nil
}

func (h *StrictHandler) DeleteWorkspace(
	ctx context.Context,
	request note.DeleteWorkspaceRequestObject,
) (note.DeleteWorkspaceResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.DeleteWorkspace{
		ID:     request.WorkspaceId,
		UserID: user.ID,
	}
	err := h.App.Cmds.DeleteWorkspace.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.DeleteWorkspace204Response{}, nil
}

func (h *StrictHandler) ChangeWorkspaceSlug(
	ctx context.Context,
	request note.ChangeWorkspaceSlugRequestObject,
) (note.ChangeWorkspaceSlugResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.ChangeWorkspaceSlug{
		ID:     request.WorkspaceId,
		Slug:   request.Body.Slug,
		UserID: user.ID,
	}
	err := h.App.Cmds.ChangeWorkspaceSlug.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.ChangeWorkspaceSlug204Response{}, nil
}

func (h *StrictHandler) GetWorkspace(
	ctx context.Context,
	request note.GetWorkspaceRequestObject,
) (note.GetWorkspaceResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	query := &app.GetWorkspaceBySlug{
		Slug:   request.WorkspaceSlug,
		UserID: user.ID,
	}
	result, err := h.App.Queries.GetWorkspace.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toWorkspaceDTO(&result)
	return note.GetWorkspace200JSONResponse(dto), nil
}

func (h *StrictHandler) CheckWorkspaceSlugExists(
	ctx context.Context,
	request note.CheckWorkspaceSlugExistsRequestObject,
) (note.CheckWorkspaceSlugExistsResponseObject, error) {
	query := &app.CheckWorkspaceSlugExists{
		Slug: request.WorkspaceSlug,
	}
	exists, err := h.App.Queries.CheckWorkspaceSlugExists.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	if exists {
		return note.CheckWorkspaceSlugExists409Response{}, nil
	}
	return note.CheckWorkspaceSlugExists200Response{}, nil
}

func (h *StrictHandler) GetMyWorkspaces(
	ctx context.Context,
	request note.GetMyWorkspacesRequestObject,
) (note.GetMyWorkspacesResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	query := &app.GetMyWorkspaces{
		UserID: user.ID,
	}
	myWorkspaces, err := h.App.Queries.GetMyWorkspaces.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	dtos := make([]note.UserWorkspace, len(myWorkspaces))
	for i := range myWorkspaces {
		dto, err := toUserWorkspaceDTO(&myWorkspaces[i])
		if err != nil {
			return nil, err
		}
		dtos[i] = dto
	}
	return note.GetMyWorkspaces200JSONResponse(dtos), nil
}

func (h *StrictHandler) GetWorkspaceEvents(
	ctx context.Context,
	request note.GetWorkspaceEventsRequestObject,
) (note.GetWorkspaceEventsResponseObject, error) {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return nil, errs.NewInternal("failed to cast context to gin.Context")
	}
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	eventCh, err := h.WorkspaceEventHub.Subscribe(ctx, &app.WorkspaceEventSubscriberParams{
		WorkspaceID: request.WorkspaceId,
		SessionID:   request.Params.SessionId,
		UserID:      user.ID,
	})
	if err != nil {
		return nil, errs.NewInternalErr("failed to subscribe to workspace events", err)
	}
	r, w := io.Pipe()
	sender := newWorkspaceEventSSESender(ctx, eventCh, w, c.Writer)
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
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	ignoreOrphans := false
	if request.Params.IncludeOrphans != nil {
		ignoreOrphans = !*request.Params.IncludeOrphans
	}
	query := &app.GetWorkspaceGraph{
		ID:            request.WorkspaceId,
		IgnoreOrphans: ignoreOrphans,
		UserID:        user.ID,
	}
	result, err := h.App.Queries.GetWorkspaceGraph.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toGraphDTO(&result)
	return note.GetWorkspaceGraph200JSONResponse(dto), nil
}

func (h *StrictHandler) LeaveWorkspace(
	ctx context.Context,
	request note.LeaveWorkspaceRequestObject,
) (note.LeaveWorkspaceResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.LeaveWorkspace{
		UserID:      user.ID,
		WorkspaceID: request.WorkspaceId,
	}
	err := h.App.Cmds.LeaveWorkspace.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.LeaveWorkspace204Response{}, nil
}

func (h *StrictHandler) GetWorkspaceMembers(
	ctx context.Context,
	request note.GetWorkspaceMembersRequestObject,
) (note.GetWorkspaceMembersResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	query := &app.GetWorkspaceMembers{
		ID:     request.WorkspaceId,
		UserID: user.ID,
	}
	result, err := h.App.Queries.GetWorkspaceMembers.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	dto, err := toWorkspaceMembersDTO(result)
	if err != nil {
		return nil, err
	}
	return note.GetWorkspaceMembers200JSONResponse(dto), nil
}

func (h *StrictHandler) UpdateWorkspaceMembers(
	ctx context.Context,
	request note.UpdateWorkspaceMembersRequestObject,
) (note.UpdateWorkspaceMembersResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	members, err := toWorkspaceMemberUpdates(*request.Body)
	if err != nil {
		return nil, err
	}
	cmd := &app.UpdateWorkspaceMembers{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
		Members:     members,
	}
	if err = h.App.Cmds.UpdateWorkspaceMembers.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return note.UpdateWorkspaceMembers204Response{}, nil
}

func (h *StrictHandler) MoveWorkspaceItems(
	ctx context.Context,
	request note.MoveWorkspaceItemsRequestObject,
) (note.MoveWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	var noteIDs []uuid.UUID
	if request.Body.NoteIds != nil {
		noteIDs = *request.Body.NoteIds
	}

	var folderIDs []uuid.UUID
	if request.Body.FolderIds != nil {
		folderIDs = *request.Body.FolderIds
	}

	destFolderID := request.Body.DestinationFolderId

	cmd := &app.MoveWorkspaceItems{
		UserID:              user.ID,
		WorkspaceID:         request.WorkspaceId,
		NoteIDs:             noteIDs,
		FolderIDs:           folderIDs,
		DestinationFolderID: destFolderID,
	}
	err := h.App.Cmds.MoveWorkspaceItems.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.MoveWorkspaceItems204Response{}, nil
}

func (h *StrictHandler) PublishWorkspace(
	ctx context.Context,
	request note.PublishWorkspaceRequestObject,
) (note.PublishWorkspaceResponseObject, error) {
	return nil, errs.Unimplemented
}

func (h *StrictHandler) RenameWorkspace(
	ctx context.Context,
	request note.RenameWorkspaceRequestObject,
) (note.RenameWorkspaceResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.RenameWorkspace{
		ID:     request.WorkspaceId,
		Name:   request.Body.Name,
		UserID: user.ID,
	}
	err := h.App.Cmds.RenameWorkspace.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.RenameWorkspace204Response{}, nil
}

func (h *StrictHandler) RestoreTrashedWorkspaceItems(
	ctx context.Context,
	request note.RestoreTrashedWorkspaceItemsRequestObject,
) (note.RestoreTrashedWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	var noteIDs []uuid.UUID
	if request.Body.NoteIds != nil {
		noteIDs = *request.Body.NoteIds
	}

	var folderIDs []uuid.UUID
	if request.Body.FolderIds != nil {
		folderIDs = *request.Body.FolderIds
	}

	cmd := &app.RestoreTrashedWorkspaceItems{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
		NoteIDs:     noteIDs,
		FolderIDs:   folderIDs,
	}
	err := h.App.Cmds.RestoreTrashedWorkspaceItems.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.RestoreTrashedWorkspaceItems204Response{}, nil
}

func (h *StrictHandler) GetWorkspaceSearchToken(ctx context.Context, request note.GetWorkspaceSearchTokenRequestObject) (note.GetWorkspaceSearchTokenResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	query := &app.GetWorkspaceSearchToken{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
	}
	result, err := h.App.Queries.GetWorkspaceSearchToken.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	dto := toSearchTokenDTO(&result)
	return note.GetWorkspaceSearchToken200JSONResponse(dto), nil
}

func (h *StrictHandler) ShowTrash(
	ctx context.Context,
	request note.ShowTrashRequestObject,
) (note.ShowTrashResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	query := &app.ShowTrash{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
	}
	result, err := h.App.Queries.ShowTrash.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto, err := toShowTrashDTO(&result)
	if err != nil {
		return nil, err
	}
	return note.ShowTrash200JSONResponse(dto), nil
}

func (h *StrictHandler) TrashWorkspaceItems(
	ctx context.Context,
	request note.TrashWorkspaceItemsRequestObject,
) (note.TrashWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.TrashWorkspaceItems{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
		NoteIDs:     *request.Body.NoteIds,
		FolderIDs:   *request.Body.FolderIds,
	}
	err := h.App.Cmds.TrashWorkspaceItems.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.TrashWorkspaceItems204Response{}, nil
}

func (h *StrictHandler) GetWorkspaceTree(
	ctx context.Context,
	request note.GetWorkspaceTreeRequestObject,
) (note.GetWorkspaceTreeResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	var depth uint
	if request.Params.Depth != nil && *request.Params.Depth > 0 {
		depth = uint(*request.Params.Depth)
	}

	var rootFolderID uuid.UUID
	if request.Params.RootFolderId != nil {
		rootFolderID = *request.Params.RootFolderId
	}

	query := &app.GetWorkspaceTree{
		WorkspaceID:    request.WorkspaceId,
		RootFolderID:   rootFolderID,
		IncludeTrashed: request.Params.IncludeTrashed != nil && *request.Params.IncludeTrashed,
		Depth:          depth,
		UserID:         user.ID,
	}

	result, err := h.App.Queries.GetWorkspaceTree.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toWorkspaceTreeFolderDTO(&result)
	return note.GetWorkspaceTree200JSONResponse(dto), nil
}

func (h *StrictHandler) UnpublishWorkspace(
	ctx context.Context,
	request note.UnpublishWorkspaceRequestObject,
) (note.UnpublishWorkspaceResponseObject, error) {
	return nil, errs.Unimplemented
}

func (h *StrictHandler) PermanentlyDeleteWorkspaceItems(
	ctx context.Context,
	request note.PermanentlyDeleteWorkspaceItemsRequestObject,
) (note.PermanentlyDeleteWorkspaceItemsResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
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
	err := h.App.Cmds.PermanentlyDeleteWorkspaceItems.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.PermanentlyDeleteWorkspaceItems204Response{}, nil
}
