package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type CreateWorkspace struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type CreateWorkspaceHandler struct {
	workspacerepo domain.WorkspaceRepo
	folderrepo    domain.FolderRepo
}

func NewCreateWorkspaceHandler(
	workspacerepo domain.WorkspaceRepo,
	folderrepo domain.FolderRepo,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		workspacerepo: workspacerepo,
		folderrepo:    folderrepo,
	}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, cmd *CreateWorkspace) error {
	// TODO: Creating a workspace requires provisioning a root folder first.
	// Suggested steps:
	// 1. Generate a rootFolderID
	// 2. Create the root Folder (with nil parentID) via FolderRepo.Save
	// 3. Create the Workspace with that rootFolderID via WorkspaceRepo.Save
	rootFolderID := uuid.New()
	rootHierarchy := domain.NewFolderHierarchy(nil)
	rootFolder := domain.NewFolder(rootFolderID, cmd.Name, nil, cmd.ID, *rootHierarchy)
	if err := h.folderrepo.Save(ctx, rootFolder); err != nil {
		return err
	}
	workspace := domain.NewWorkspace(cmd.ID, cmd.Name, cmd.Slug, rootFolderID)
	return h.workspacerepo.Save(ctx, workspace)
}
