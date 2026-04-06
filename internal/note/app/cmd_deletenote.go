package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type PermanentlyDeleteNote struct {
	ID     uuid.UUID
	UserID string
}

type PermanentlyDeleteNoteHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
}

func PermanentlyNewDeleteNoteHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
) *PermanentlyDeleteNoteHandler {
	return &PermanentlyDeleteNoteHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
	}
}

var ProvidePermanentlyDeleteNoteHandler = PermanentlyNewDeleteNoteHandler

func (h *PermanentlyDeleteNoteHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteNote) error {
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
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to delete note %s", cmd.UserID, cmd.ID),
		)
	}

	return h.noteRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}

var ErrCodeDeleteNoteForbidden = "DeleteNote_1"
