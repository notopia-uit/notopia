package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type DeleteFolder struct {
	ID     uuid.UUID
	UserID string
}

type DeleteFolderHandler struct {
	authorizationService AuthorizationService
	folderRepo           domain.FolderRepo
}

func NewDeleteFolderHandler(
	authorizationService AuthorizationService,
	folderRepo domain.FolderRepo,
) *DeleteFolderHandler {
	return &DeleteFolderHandler{
		authorizationService: authorizationService,
		folderRepo:           folderRepo,
	}
}

var ProvideDeleteFolderHandler = NewDeleteFolderHandler

func (h *DeleteFolderHandler) Handle(ctx context.Context, cmd *DeleteFolder) error {
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
		return newErrDeleteFolderForbidden(cmd.UserID, workspaceID)
	}

	return h.folderRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}

var ErrCodeDeleteFolderForbidden = "DeleteFolder_1"

func newErrDeleteFolderForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to delete folder in workspace %q", userID, workspaceID.String()),
		ErrCodeDeleteFolderForbidden,
		nil,
	)
}
