package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type DeleteFolder struct {
	ID     uuid.UUID
	UserID string
}

type DeleteFolderHandler struct {
	authorization service.Authorization
	folderRepo    domain.FolderRepo
}

func NewDeleteFolderHandler(
	authorization service.Authorization,
	folderRepo domain.FolderRepo,
) *DeleteFolderHandler {
	return &DeleteFolderHandler{
		authorization: authorization,
		folderRepo:    folderRepo,
	}
}

var ProvideDeleteFolderHandler = NewDeleteFolderHandler

func (h *DeleteFolderHandler) Handle(ctx context.Context, cmd *DeleteFolder) errs.Error {
	workspaceID, err := h.folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		service.WorkspaceItemPermissionDelete,
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
