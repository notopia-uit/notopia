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
	folderRepo           domain.FolderRepo
}

func PermanentlyNewDeleteFolderHandler(
	authorizationService AuthorizationService,
	folderRepo domain.FolderRepo,
) *PermanentlyDeleteFolderHandler {
	return &PermanentlyDeleteFolderHandler{
		authorizationService: authorizationService,
		folderRepo:           folderRepo,
	}
}

var ProvidePermanentlyDeleteFolderHandler = PermanentlyNewDeleteFolderHandler

func (h *PermanentlyDeleteFolderHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteFolder) errs.Error {
	workspaceID, err := h.folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
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
			fmt.Sprintf("user %s does not have permission to delete folder %s", cmd.UserID, cmd.ID),
		)
	}

	return h.folderRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}
