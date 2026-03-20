package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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
	workspaceEventPubSub WorkspaceEventPubSub
}

func NewCreateFolderHandler(
	authorizationService AuthorizationService,
	folderRepo domain.FolderRepo,
	workspaceEventPubSub WorkspaceEventPubSub,
) *CreateFolderHandler {
	return &CreateFolderHandler{
		authorizationService: authorizationService,
		folderRepo:           folderRepo,
		workspaceEventPubSub: workspaceEventPubSub,
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
		return newErrCreateFolderForbidden(cmd.UserID, cmd.WorkspaceID)
	}
	hierarchy := domain.NewFolderHierarchy(&cmd.ParentID)
	folder, err := domain.NewFolder(cmd.ID, cmd.Name, cmd.Icon, cmd.WorkspaceID, *hierarchy)
	if err != nil {
		return err
	}
	if err := h.folderRepo.Save(ctx, folder); err != nil {
		return err
	}
	return h.workspaceEventPubSub.Publish(ctx, cmd.WorkspaceID, cmd.UserID, folder.PopEvents()...)
}

var ErrCodeCreateFolderForbidden = "CreateFolder_1"

func newErrCreateFolderForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to create folder in workspace %q", userID, workspaceID.String()),
		ErrCodeCreateFolderForbidden,
		nil,
	)
}
