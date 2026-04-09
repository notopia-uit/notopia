package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type CreateWorkspace struct {
	ID      uuid.UUID
	Name    string
	Slug    string
	OwnerID string
}

type CreateWorkspaceHandler struct {
	uow                  domain.UnitOfWork
	authorizationService AuthorizationService
}

func NewCreateWorkspaceHandler(
	uow domain.UnitOfWork,
	authorizationService AuthorizationService,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		uow:                  uow,
		authorizationService: authorizationService,
	}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, cmd *CreateWorkspace) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		workspaceRepo := r.Workspace()
		folderRepo := r.Folder()
		slugExisted, err := workspaceRepo.CheckSlugExists(ctx, cmd.Slug)
		if err != nil {
			return err
		}
		if slugExisted {
			return errs.NewWorkspaceSlugAlreadyExists(cmd.Slug, nil)
		}
		rootFolderID, err := uuid.NewV7()
		if err != nil {
			return errs.NewInternalGenerateID(err)
		}
		rootFolder, err := domain.NewFolder(rootFolderID, cmd.Name, "", cmd.ID, domain.FolderHierarchy{}, cmd.OwnerID)
		if err != nil {
			return err
		}
		workspace, err := domain.NewWorkspace(cmd.ID, cmd.Name, cmd.Slug, rootFolderID)
		if err != nil {
			return err
		}
		if err := folderRepo.Save(ctx, rootFolder); err != nil {
			return err
		}
		if err := workspaceRepo.Save(ctx, workspace); err != nil {
			return err
		}
		if err := h.authorizationService.CreateWorkspaceWithOwner(ctx, cmd.OwnerID, workspace.ID()); err != nil {
			return err
		}
		return nil
	})
}
