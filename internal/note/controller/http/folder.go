package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func (h *StrictHandler) CreateFolder(
	ctx context.Context,
	request note.CreateFolderRequestObject,
) (note.CreateFolderResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	body := request.Body

	id, err := uuid.NewV7()
	if err != nil {
		return nil, errs.NewInternalGenerateID(err)
	}
	icon, _ := body.Icon.Get()
	parentID, _ := body.ParentId.Get()
	cmd := &app.CreateFolder{
		ID:          id,
		Name:        body.Name,
		Icon:        icon,
		ParentID:    parentID,
		WorkspaceID: body.WorkspaceId,
		UserID:      user.ID,
	}

	if err := h.App.Cmds.CreateFolder.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return note.CreateFolder201JSONResponse{
		Headers: note.CreateFolder201ResponseHeaders{
			ContentLocation: h.BaseURL.JoinPath("note", "folders", id.String()).String(),
		},
		Body: note.CreateFolderResponse{
			Id: &id,
		},
	}, nil
}

func (h *StrictHandler) PermanentlyDeleteFolder(
	ctx context.Context,
	request note.PermanentlyDeleteFolderRequestObject,
) (note.PermanentlyDeleteFolderResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.PermanentlyDeleteFolder{
		ID:     request.FolderId,
		UserID: user.ID,
	}

	if err := h.App.Cmds.DeleteFolder.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return note.PermanentlyDeleteFolder204Response{}, nil
}

func (h *StrictHandler) RenameFolder(
	ctx context.Context,
	request note.RenameFolderRequestObject,
) (note.RenameFolderResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	body := request.Body

	cmd := &app.RenameFolder{
		ID:     request.FolderId,
		Name:   body.Name,
		UserID: user.ID,
	}

	if err := h.App.Cmds.RenameFolder.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return note.RenameFolder204Response{}, nil
}
