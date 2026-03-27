package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameNote struct {
	ID     uuid.UUID
	Name   string
	UserID string
}

type RenameNoteHandler struct {
	authorization service.Authorization
	noteRepo      domain.NoteRepo
}

func NewRenameNoteHandler(
	authorization service.Authorization,
	noteRepo domain.NoteRepo,
) *RenameNoteHandler {
	return &RenameNoteHandler{
		authorization: authorization,
		noteRepo:      noteRepo,
	}
}

var ProvideRenameNoteHandler = NewRenameNoteHandler

func (h *RenameNoteHandler) Handle(ctx context.Context, cmd *RenameNote) errs.Error {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		service.WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to rename note %s", cmd.UserID, cmd.ID),
		)
	}

	note, err := h.noteRepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	note.Rename(cmd.Name)
	return h.noteRepo.Save(ctx, note)
}
