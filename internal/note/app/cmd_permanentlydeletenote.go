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
	uow                  domain.UnitOfWork
}

func PermanentlyNewDeleteNoteHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
) *PermanentlyDeleteNoteHandler {
	return &PermanentlyDeleteNoteHandler{
		authorizationService: authorizationService,
		uow:                  uow,
	}
}

var ProvidePermanentlyDeleteNoteHandler = PermanentlyNewDeleteNoteHandler

func (h *PermanentlyDeleteNoteHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteNote) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		noteRepo := r.Note()
		workspaceID, err := noteRepo.GetWorkspaceIDByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		// NOTE: @coderabbitai
		//	Avoid holding the delete transaction open during authorization.
		//	This refactor puts HasWorkspaceItemPermission(...) inside uow.Execute(...).
		//	If the authorization check is remote, the delete transaction stays open while waiting on another service,
		//	which increases lock time and failure blast radius for a simple permission lookup.
		//	Keep the auth check outside the write transaction, then re-load/delete inside the transaction.
		hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(ctx, cmd.UserID, workspaceID, WorkspaceItemPermissionDelete)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errs.NewForbidden(
				fmt.Sprintf("user %s does not have permission to delete note %s", cmd.UserID, cmd.ID),
			)
		}
		note, err := noteRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		note.PermanentlyDelete(cmd.UserID)
		return noteRepo.Save(ctx, note)
	})
}
