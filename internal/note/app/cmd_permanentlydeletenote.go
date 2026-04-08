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
	uow                  domain.UnitOfWork
}

func PermanentlyNewDeleteNoteHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	uow domain.UnitOfWork,
) *PermanentlyDeleteNoteHandler {
	return &PermanentlyDeleteNoteHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		uow:                  uow,
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

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		noteRepo := r.Note()
		note, err := noteRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		note.Deleted()
		return noteRepo.Save(ctx, note)
	})
}
