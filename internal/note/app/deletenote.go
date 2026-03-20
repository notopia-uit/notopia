package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type DeleteNote struct {
	ID     uuid.UUID
	UserID string
}

type DeleteNoteHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
}

func NewDeleteNoteHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
) *DeleteNoteHandler {
	return &DeleteNoteHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
	}
}

var ProvideDeleteNoteHandler = NewDeleteNoteHandler

func (h *DeleteNoteHandler) Handle(ctx context.Context, cmd *DeleteNote) error {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		WorkspaceItemPermissionDelete,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return newErrDeleteNoteForbidden(cmd.UserID, workspaceID)
	}

	return h.noteRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}

var ErrCodeDeleteNoteForbidden = "DeleteNote_1"

func newErrDeleteNoteForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to delete note in workspace %q", userID, workspaceID.String()),
		ErrCodeDeleteNoteForbidden,
		nil,
	)
}
