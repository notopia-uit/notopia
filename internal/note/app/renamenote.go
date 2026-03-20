package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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

func (h *RenameNoteHandler) Handle(ctx context.Context, cmd *RenameNote) error {
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
		return newErrRenameNoteForbidden(cmd.UserID, workspaceID)
	}

	note, err := h.noterepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return domain.NewErrNoteNotFound(cmd.ID, err)
	}
	note.Rename(cmd.Name)
	return h.noterepo.Save(ctx, note)
}

var ErrCodeRenameNoteForbidden = "RenameNote_1"

func newErrRenameNoteForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to rename note in workspace %q", userID, workspaceID.String()),
		ErrCodeRenameNoteForbidden,
		nil,
	)
}
