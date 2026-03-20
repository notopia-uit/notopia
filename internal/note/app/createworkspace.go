package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type CreateWorkspace struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type CreateWorkspaceHandler struct {
	workspaceRepo domain.WorkspaceRepo
	folderRepo    domain.FolderRepo
	uow           domain.UnitOfWork
}

func NewCreateWorkspaceHandler(
	workspaceRepo domain.WorkspaceRepo,
	folderRepo domain.FolderRepo,
	uow domain.UnitOfWork,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		workspaceRepo: workspaceRepo,
		folderRepo:    folderRepo,
		uow:           uow,
	}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, cmd *CreateWorkspace) error {
	slugExisted, err := h.workspaceRepo.CheckSlugExists(ctx, cmd.Slug)
	if err != nil {
		return err
	}
	if slugExisted {
		return newErrCreateWorkspaceSlugExisted(cmd.Slug)
	}
	rootFolderID := uuid.New()
	rootHierarchy := domain.NewFolderHierarchy(nil)
	rootFolder, err := domain.NewFolder(rootFolderID, cmd.Name, nil, cmd.ID, *rootHierarchy)
	if err != nil {
		return err
	}
	workspace, err := domain.NewWorkspace(cmd.ID, cmd.Name, cmd.Slug, rootFolderID)
	if err != nil {
		return err
	}
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if err := h.folderRepo.Save(ctx, rootFolder); err != nil {
			return err
		}
		if err := h.workspaceRepo.Save(ctx, workspace); err != nil {
			return err
		}
		return nil
	})
	return err
}

var ErrCodeCreateWorkspaceSlugExisted = "ErrCodeCreateWorkspaceSlugExisted"

func newErrCreateWorkspaceSlugExisted(slug string) *commonerror.Err {
	return commonerror.NewInvalid(
		fmt.Sprintf("workspace slug %q already exists", slug),
		ErrCodeCreateWorkspaceSlugExisted,
		nil,
	)
}
