package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type DeleteFolder struct {
	ID uuid.UUID
}

type DeleteFolderHandler struct {
	folderRepo domain.FolderRepo
}

func NewDeleteFolderHandler(folderRepo domain.FolderRepo) *DeleteFolderHandler {
	return &DeleteFolderHandler{folderRepo: folderRepo}
}

var ProvideDeleteFolderHandler = NewDeleteFolderHandler

func (h *DeleteFolderHandler) Handle(ctx context.Context, cmd *DeleteFolder) error {
	return h.folderRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}
