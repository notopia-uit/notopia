package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameNote struct {
	ID     uuid.UUID
	Name   string
	UserID string
}

type RenameNoteHandler struct {
	authorizationService AuthorizationService
	noterepo             domain.NoteRepo
}

func NewRenameNoteHandler(
	authorizationService AuthorizationService,
	noterepo domain.NoteRepo,
) *RenameNoteHandler {
	return &RenameNoteHandler{
		authorizationService: authorizationService,
		noterepo:             noterepo,
	}
}

var ProvideRenameNoteHandler = NewRenameNoteHandler

func (h *RenameNoteHandler) Handle(ctx context.Context, cmd *RenameNote) errs.Error {
	workspaceID, err := h.noterepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to rename note %s", cmd.UserID, cmd.ID),
		)
	}

	note, err := h.noterepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	note.Rename(cmd.Name)
	return h.noterepo.Save(ctx, note)
}
