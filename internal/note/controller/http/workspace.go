package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
	"time"

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
		return nil, errs.NewInternal("failed to generate UUIDv7 for new workspace", err)
	}
	cmd := &app.CreateWorkspace{
		ID:   id,
		Name: request.Body.Name,
		Slug: request.Body.Slug,
	}
	err = h.App.CreateWorkspaceHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateWorkspace201Response{
		Headers: note.CreateWorkspace201ResponseHeaders{
			ContentLocation: h.ServerURL + "/workspaces/" + id.String(),
		},
	}, nil
}

func (h *StrictHandler) DeleteWorkspace(
	ctx context.Context,
	request note.DeleteWorkspaceRequestObject,
) (note.DeleteWorkspaceResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &app.DeleteWorkspace{
		ID:     request.WorkspaceId,
		UserID: user.ID,
	}
	err = h.App.DeleteWorkspaceHandler.Handle(ctx, cmd)
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
	result, err := h.App.GetWorkspaceHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getWorkspaceToDTO(*result)
	return note.GetWorkspace200JSONResponse(dto), nil
}

func (h *StrictHandler) CheckWorkspaceSlugExists(
	ctx context.Context,
	request note.CheckWorkspaceSlugExistsRequestObject,
) (note.CheckWorkspaceSlugExistsResponseObject, error) {
	query := &app.CheckWorkspaceSlugExists{
		Slug: request.WorkspaceSlug,
	}
	result, err := h.App.CheckWorkspaceSlugExistsHandler.Handle(ctx, query)
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
	user, err := commonhttp.UserFromContextError(c)
	if err != nil {
		return nil, err
	}
	eventCh, err := h.WorkspaceEventPubSub.Subscribe(ctx, request.WorkspaceId, user.ID)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	r, w := io.Pipe()
	mu := &sync.Mutex{}
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
				mu.Lock()
				if _, err := w.Write([]byte("heartbeat: keep-alive\n\n")); err != nil {
					slog.ErrorContext(c, "failed to write keep-alive comment in workspace events stream", slog.String("error", err.Error()))
					mu.Unlock()
					return
				}
				c.Writer.Flush()
				mu.Unlock()
				slog.DebugContext(c, "sent keep-alive comment in workspace events stream")
			case event, ok := <-eventCh:
				if !ok {
					slog.InfoContext(c, "workspace event channel closed")
					return
				}
				dto, ok := workspaceEventToDTO(event)
				if !ok {
					slog.WarnContext(c, "skipping unsupported workspace event type in workspace events stream")
					continue
				}
				eventBytes, err := json.Marshal(dto)
				if err != nil {
					slog.ErrorContext(c, "failed to marshal event to JSON", slog.String("error", err.Error()))
					continue
				}
				mu.Lock()
				if _, err := w.Write([]byte("data: ")); err != nil {
					slog.ErrorContext(c, "failed to write event prefix in workspace events stream", slog.String("error", err.Error()))
					mu.Unlock()
					return
				}
				if _, err := w.Write(eventBytes); err != nil {
					slog.ErrorContext(c, "failed to write event data in workspace events stream", slog.String("error", err.Error()))
					mu.Unlock()
					return
				}
				if _, err := w.Write([]byte("\n\n")); err != nil {
					slog.ErrorContext(c, "failed to write event suffix in workspace events stream", slog.String("error", err.Error()))
					mu.Unlock()
					return
				}
				c.Writer.Flush()
				mu.Unlock()
			}
		}
	}()
	//exhaustruct:ignore
	return note.GetWorkspaceEvents200TexteventStreamResponse{
		Body: r,
	}, nil
}

func (h *StrictHandler) GetWorkspaceGraph(
	ctx context.Context,
	request note.GetWorkspaceGraphRequestObject,
) (note.GetWorkspaceGraphResponseObject, error) {
	query := &app.GetWorkspaceGraph{
		ID:     request.WorkspaceId,
		Orphan: request.Params.Orphan,
	}
	result, err := h.App.GetWorkspaceGraphHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getGraphToDTO(result)
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
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	var noteIDs []uuid.UUID
	if request.Body.NoteIds != nil {
		noteIDs = make([]uuid.UUID, len(*request.Body.NoteIds))
		copy(noteIDs, *request.Body.NoteIds)
	}

	var folderIDs []uuid.UUID
	if request.Body.FolderIds != nil {
		folderIDs = make([]uuid.UUID, len(*request.Body.FolderIds))
		copy(folderIDs, *request.Body.FolderIds)
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
	err = h.App.MoveWorkspaceItemsHandler.Handle(ctx, cmd)
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
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &app.RenameWorkspace{
		ID:     request.WorkspaceId,
		Name:   request.Body.Name,
		UserID: user.ID,
	}
	err = h.App.RenameWorkspaceHandler.Handle(ctx, cmd)
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
	result, err := h.App.ShowTrashHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getTrashedToDTO(result)
	return note.ShowTrash200JSONResponse(dto), nil
}

func (h *StrictHandler) TrashWorkspaceItems(
	ctx context.Context,
	request note.TrashWorkspaceItemsRequestObject,
) (note.TrashWorkspaceItemsResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	var noteIDs []uuid.UUID
	if request.Body.Notes != nil {
		noteIDs = make([]uuid.UUID, len(*request.Body.Notes))
		for i, item := range *request.Body.Notes {
			noteIDs[i] = item.Id
		}
	}

	var folderIDs []uuid.UUID
	if request.Body.Folders != nil {
		folderIDs = make([]uuid.UUID, len(*request.Body.Folders))
		for i, item := range *request.Body.Folders {
			folderIDs[i] = item.Id
		}
	}

	cmd := &app.TrashWorkspaceItems{
		WorkspaceID: request.WorkspaceId,
		UserID:      user.ID,
		NoteIDs:     noteIDs,
		FolderIDs:   folderIDs,
	}
	err = h.App.TrashWorkspaceItemsHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.TrashWorkspaceItems204Response{}, nil
}

func (h *StrictHandler) GetWorkspaceTree(
	ctx context.Context,
	request note.GetWorkspaceTreeRequestObject,
) (note.GetWorkspaceTreeResponseObject, error) {
	query := &app.GetWorkspaceTree{
		ID: request.WorkspaceId,
	}
	result, err := h.App.GetWorkspaceTreeHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getWorkspaceTreeFolderToDTO(result)
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
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
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
	err = h.App.PermanentlyDeleteWorkspaceItemsHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.PermanentlyDeleteWorkspaceItems204Response{}, nil
}
