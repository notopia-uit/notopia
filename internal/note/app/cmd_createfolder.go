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
	Icon        string
	ParentID    uuid.UUID
	WorkspaceID uuid.UUID

	UserID string
}
type CreateFolderHandler struct {
	authorizationSvc AuthorizationSvc
	uow              domain.UnitOfWork
}

func NewCreateFolderHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *CreateFolderHandler {
	return &CreateFolderHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
	}
}

var ProvideCreateFolderHandler = NewCreateFolderHandler

func (h *CreateFolderHandler) Handle(ctx context.Context, cmd *CreateFolder) error {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionWrite)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %q does not have permission to create folder in workspace %q", cmd.UserID, cmd.WorkspaceID.String()),
		)
	}
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		exist, err := folderRepo.CheckExists(ctx, cmd.ID)
		if err != nil {
			return err
		}
		if exist {
			return errs.NewFolderAlreadyExisted(cmd.ID)
		}
		hierarchy := domain.NewFolderHierarchy(cmd.ParentID)
		folder, err := domain.NewFolder(cmd.ID, cmd.Name, cmd.Icon, cmd.WorkspaceID, hierarchy, cmd.UserID)
		if err != nil {
			return err
		}
		return folderRepo.Save(ctx, folder)
	})
}
