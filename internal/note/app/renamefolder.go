package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type RenameFolder struct {
	ID   uuid.UUID
	Name string
}

type RenameFolderHandler struct {
	folderrepo domain.FolderRepo
}

func NewRenameFolderHandler(folderrepo domain.FolderRepo) *RenameFolderHandler {
	return &RenameFolderHandler{folderrepo: folderrepo}
}

var ProvideRenameFolderHandler = NewRenameFolderHandler

func (h *RenameFolderHandler) Handle(ctx context.Context, cmd *RenameFolder) error {
	folder, err := h.folderrepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return domain.NewErrFolderNotFound(cmd.ID, err)
	}
	folder.Rename(cmd.Name)
	return h.folderrepo.Save(ctx, folder)
}
