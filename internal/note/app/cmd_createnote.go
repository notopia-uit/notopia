package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type CreateNote struct {
	ID       uuid.UUID
	Name     string
	Icon     string
	FolderID uuid.UUID

	UserID string
}

type CreateNoteHandler struct {
	authorizationService AuthorizationService
	uow                  domain.UnitOfWork
}

func NewCreateNoteHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
) *CreateNoteHandler {
	return &CreateNoteHandler{
		authorizationService: authorizationService,
		uow:                  uow,
	}
}

var ProvideCreateNoteHandler = NewCreateNoteHandler

func (h *CreateNoteHandler) Handle(ctx context.Context, cmd *CreateNote) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		noteRepo := r.Note()
		folderExists, err := folderRepo.CheckExists(ctx, cmd.FolderID)
		if err != nil {
			return err
		}
		if !folderExists {
			return errs.NewFolderNotFound(cmd.FolderID, err)
		}
		workspaceID, err := folderRepo.GetWorkspaceIDByID(ctx, cmd.FolderID)
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
				fmt.Sprintf("user %q does not have permission to create note in workspace %q", cmd.UserID, workspaceID.String()),
			)
		}
		note := domain.NewNote(cmd.ID, cmd.Name, cmd.Icon, cmd.FolderID)
		return noteRepo.Save(ctx, note)
	})
}
