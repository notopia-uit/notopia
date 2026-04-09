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
	uow                  domain.UnitOfWork
}

func NewRenameNoteHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
) *RenameNoteHandler {
	return &RenameNoteHandler{
		authorizationService: authorizationService,
		uow:                  uow,
	}
}

var ProvideRenameNoteHandler = NewRenameNoteHandler

func (h *RenameNoteHandler) Handle(ctx context.Context, cmd *RenameNote) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		noteRepo := r.Note()
		workspaceID, err := noteRepo.GetWorkspaceIDByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(ctx, cmd.UserID, workspaceID, WorkspaceItemPermissionWrite)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errs.NewForbidden(
				fmt.Sprintf("user %s does not have permission to rename note %s", cmd.UserID, cmd.ID),
			)
		}
		note, err := noteRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		note.Rename(cmd.Name, cmd.UserID)
		return noteRepo.Save(ctx, note)
	})
}
