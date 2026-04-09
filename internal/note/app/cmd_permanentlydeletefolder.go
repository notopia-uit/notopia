package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type PermanentlyDeleteFolder struct {
	ID     uuid.UUID
	UserID string
}

type PermanentlyDeleteFolderHandler struct {
	authorizationService AuthorizationService
	uow                  domain.UnitOfWork
}

func PermanentlyNewDeleteFolderHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
) *PermanentlyDeleteFolderHandler {
	return &PermanentlyDeleteFolderHandler{
		authorizationService: authorizationService,
		uow:                  uow,
	}
}

var ProvidePermanentlyDeleteFolderHandler = PermanentlyNewDeleteFolderHandler

// NOTE: We delegate the infra persistence to cascading delete things
// Fact, we should handle this in domain, not infra
func (h *PermanentlyDeleteFolderHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteFolder) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		workspaceID, err := folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(ctx, cmd.UserID, workspaceID, WorkspaceItemPermissionDelete)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errs.NewForbidden(
				fmt.Sprintf("user %s does not have permission to delete folder %s", cmd.UserID, cmd.ID),
			)
		}
		folder, err := folderRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		folder.PermanentlyDelete(cmd.UserID)
		return folderRepo.Save(ctx, folder)
	})
}
