package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type CreateFolder struct {
	ID          uuid.UUID
	Name        string
	Icon        *string
	ParentID    uuid.UUID
	WorkspaceID uuid.UUID

	UserID string
}

type CreateFolderHandler struct {
	authorizationService AuthorizationService
	folderRepo           domain.FolderRepo
}

func NewCreateFolderHandler(
	authorizationService AuthorizationService,
	folderRepo domain.FolderRepo,
) *CreateFolderHandler {
	return &CreateFolderHandler{
		authorizationService: authorizationService,
		folderRepo:           folderRepo,
	}
}

var ProvideCreateFolderHandler = NewCreateFolderHandler

func (h *CreateFolderHandler) Handle(ctx context.Context, cmd *CreateFolder) error {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %q does not have permission to create folder in workspace %q", cmd.UserID, cmd.WorkspaceID.String()),
		)
	}
	hierarchy := domain.NewFolderHierarchy(&cmd.ParentID)
	folder, err := domain.NewFolder(cmd.ID, cmd.Name, cmd.Icon, cmd.WorkspaceID, *hierarchy, cmd.UserID)
	if err != nil {
		return err
	}
	if err := h.folderRepo.Save(ctx, folder); err != nil {
		return err
	}
	return nil
}
