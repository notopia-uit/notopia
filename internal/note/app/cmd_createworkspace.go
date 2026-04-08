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
	workspaceRepo        domain.WorkspaceRepo
	folderRepo           domain.FolderRepo
	uow                  domain.UnitOfWork
	authorizationService AuthorizationService
}

func NewCreateWorkspaceHandler(
	workspaceRepo domain.WorkspaceRepo,
	folderRepo domain.FolderRepo,
	uow domain.UnitOfWork,
	authorizationService AuthorizationService,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		workspaceRepo:        workspaceRepo,
		folderRepo:           folderRepo,
		uow:                  uow,
		authorizationService: authorizationService,
	}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, cmd *CreateWorkspace) error {
	slugExisted, err := h.workspaceRepo.CheckSlugExists(ctx, cmd.Slug)
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
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if err := h.folderRepo.Save(ctx, rootFolder); err != nil {
			return err
		}
		if err := h.workspaceRepo.Save(ctx, workspace); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if err := h.authorizationService.CreateWorkspaceWithOwnership(ctx, cmd.OwnerID, workspace.ID()); err != nil {
		return err
	}
	return err
}
