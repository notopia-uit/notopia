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
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		WorkspaceItemPermissionDelete,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to permanently delete items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	// TODO: first, need to fix the param back, awful

	// return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
	// 	folderRepo := r.Folder()
	// 	folder, err := folderRepo.GetMany(ctx, domain.NewFolderRepoGetManyParamsByIDs(cmd.FolderIDs).WithTrashed())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	folder.Deleted()
	// 	return folderRepo.Save(ctx, folder)
	// })

	// if len(cmd.NoteIDs) > 0 {
	// 	if err := h.noteRepo.PermanentlyDeleteByIDs(ctx, cmd.NoteIDs); err != nil {
	// 		return err
	// 	}
	// }
	//
	// if len(cmd.FolderIDs) > 0 {
	// 	if err := h.folderRepo.PermanentlyDeleteByIDs(ctx, cmd.FolderIDs); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
