package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/notopia-uit/notopia/internal/note/app/service"
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
	authorization service.Authorization
	folderRepo    domain.FolderRepo
	eventPubSub   pubsub.WorkspaceEvent
}

func NewCreateFolderHandler(
	authorization service.Authorization,
	folderRepo domain.FolderRepo,
	eventPubSub pubsub.WorkspaceEvent,
) *CreateFolderHandler {
	return &CreateFolderHandler{
		authorization: authorization,
		folderRepo:    folderRepo,
		eventPubSub:   eventPubSub,
	}
}

var ProvideCreateFolderHandler = NewCreateFolderHandler

func (h *CreateFolderHandler) Handle(ctx context.Context, cmd *CreateFolder) errs.Error {
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		service.WorkspaceItemPermissionWrite,
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
	folder, err := domain.NewFolder(cmd.ID, cmd.Name, cmd.Icon, cmd.WorkspaceID, *hierarchy)
	if err != nil {
		return err
	}
	if err := h.folderRepo.Save(ctx, folder); err != nil {
		return err
	}
	return h.eventPubSub.Publish(ctx, cmd.WorkspaceID, cmd.UserID, folder.PopEvents()...)
}
