package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func (h *StrictHandler) CreateFolder(
	ctx context.Context,
	request note.CreateFolderRequestObject,
) (note.CreateFolderResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	body := request.Body

	id := uuid.New()
	cmd := &app.CreateFolder{
		ID:          id,
		Name:        body.Name,
		Icon:        body.Icon,
		ParentID:    *body.ParentId,
		WorkspaceID: *body.WorkspaceId,
		UserID:      user.ID,
	}

	if err := h.App.CreateFolderHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return note.CreateFolder201Response{
		Headers: note.CreateFolder201ResponseHeaders{
			ContentLocation: h.ServerURL + "/folders/" + id.String(),
		},
	}, nil
}

func (h *StrictHandler) DeleteFolder(
	ctx context.Context,
	request note.DeleteFolderRequestObject,
) (note.DeleteFolderResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &app.DeleteFolder{
		ID:     request.FolderId,
		UserID: user.ID,
	}

	if err := h.App.DeleteFolderHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return note.DeleteFolder204Response{}, nil
}

func (h *StrictHandler) RenameFolder(
	ctx context.Context,
	request note.RenameFolderRequestObject,
) (note.RenameFolderResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	body := request.Body

	cmd := &app.RenameFolder{
		ID:     request.FolderId,
		Name:   body.Name,
		UserID: user.ID,
	}

	if err := h.App.RenameFolderHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return note.RenameFolder204Response{}, nil
}
