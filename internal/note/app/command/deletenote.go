package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type DeleteNote struct {
	ID     uuid.UUID
	UserID string
}

type DeleteNoteHandler struct {
	authorization service.Authorization
	noteRepo      domain.NoteRepo
}

func NewDeleteNoteHandler(
	authorization service.Authorization,
	noteRepo domain.NoteRepo,
) *DeleteNoteHandler {
	return &DeleteNoteHandler{
		authorization: authorization,
		noteRepo:      noteRepo,
	}
}

var ProvideDeleteNoteHandler = NewDeleteNoteHandler

func (h *DeleteNoteHandler) Handle(ctx context.Context, cmd *DeleteNote) errs.Error {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		service.WorkspaceItemPermissionDelete,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to delete note %s", cmd.UserID, cmd.ID),
		)
	}

	return h.noteRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}

var ErrCodeDeleteNoteForbidden = "DeleteNote_1"
