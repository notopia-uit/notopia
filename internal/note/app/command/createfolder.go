package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type CreateFolder struct {
	ID          uuid.UUID
	Name        string
	Icon        *string
	ParentID    uuid.UUID
	WorkspaceID uuid.UUID
}

type CreateFolderHandler struct {
	folderrepo domain.FolderRepo
}

func NewCreateFolderHandler(folderrepo domain.FolderRepo) *CreateFolderHandler {
	return &CreateFolderHandler{folderrepo: folderrepo}
}

func (h *CreateFolderHandler) Handle(ctx context.Context, cmd *CreateFolder) error {
	hierarchy := domain.NewFolderHierarchy(&cmd.ParentID)
	folder := domain.NewFolder(cmd.ID, cmd.Name, cmd.Icon, cmd.WorkspaceID, *hierarchy)
	return h.folderrepo.Save(ctx, folder)
}
