package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type RenameFolder struct {
	ID     uuid.UUID
	Name   string
	UserID string
}

type RenameFolderHandler struct {
	authorizationService AuthorizationService
	folderrepo           domain.FolderRepo
}

func NewRenameFolderHandler(
	authorizationService AuthorizationService,
	folderrepo domain.FolderRepo,
) *RenameFolderHandler {
	return &RenameFolderHandler{
		authorizationService: authorizationService,
		folderrepo:           folderrepo,
	}
}

var ProvideRenameFolderHandler = NewRenameFolderHandler

func (h *RenameFolderHandler) Handle(ctx context.Context, cmd *RenameFolder) error {
	workspaceID, err := h.folderrepo.GetWorkspaceIDByID(ctx, cmd.ID)
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
		return newErrRenameFolderForbidden(cmd.UserID, workspaceID)
	}

	folder, err := h.folderrepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return domain.NewErrFolderNotFound(cmd.ID, err)
	}
	folder.Rename(cmd.Name)
	return h.folderrepo.Save(ctx, folder)
}

var ErrCodeRenameFolderForbidden = "RenameFolder_1"

func newErrRenameFolderForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to rename folder in workspace %q", userID, workspaceID.String()),
		ErrCodeRenameFolderForbidden,
		nil,
	)
}
