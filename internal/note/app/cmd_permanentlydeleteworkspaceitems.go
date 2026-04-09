package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type PermanentlyDeleteWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type PermanentlyDeleteWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	uow                  domain.UnitOfWork
}

func NewPermanentlyDeleteWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
) *PermanentlyDeleteWorkspaceItemsHandler {
	return &PermanentlyDeleteWorkspaceItemsHandler{
		authorizationService: authorizationService,
		uow:                  uow,
	}
}

var ProvidePermanentlyDeleteWorkspaceItemsHandler = NewPermanentlyDeleteWorkspaceItemsHandler

func (h *PermanentlyDeleteWorkspaceItemsHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteWorkspaceItems) error {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionDelete)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to permanently delete items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if len(cmd.FolderIDs) > 0 {
			folderRepo := r.Folder()
			folders, err := folderRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.FolderRepoGetManyParams{
					IDs:       cmd.FolderIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}
			for _, folder := range folders {
				folder.PermanentlyDelete(cmd.UserID)
			}
			if err := folderRepo.SaveMany(ctx, folders); err != nil {
				return err
			}
		}

		if len(cmd.NoteIDs) > 0 {
			noteRepo := r.Note()
			notes, err := noteRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.NoteRepoGetManyParams{
					IDs:       cmd.NoteIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}
			for _, note := range notes {
				note.PermanentlyDelete(cmd.UserID)
			}
			if err := noteRepo.SaveMany(ctx, notes); err != nil {
				return err
			}
		}
		return nil
	})
}
