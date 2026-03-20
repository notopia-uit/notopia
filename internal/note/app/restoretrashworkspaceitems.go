package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type RestoreTrashedWorkspaceItems struct {
	WorkspaceSlug string
	NoteIDs       []uuid.UUID
	FolderIDs     []uuid.UUID
}

type RestoreTrashedWorkspaceItemsHandler struct {
	noterepo   domain.NoteRepo
	folderrepo domain.FolderRepo
}

func NewRestoreTrashedWorkspaceItemsHandler(
	noterepo domain.NoteRepo,
	folderrepo domain.FolderRepo,
) *RestoreTrashedWorkspaceItemsHandler {
	return &RestoreTrashedWorkspaceItemsHandler{
		noterepo:   noterepo,
		folderrepo: folderrepo,
	}
}

var ProvideRestoreTrashedWorkspaceItemsHandler = NewRestoreTrashedWorkspaceItemsHandler

func (h *RestoreTrashedWorkspaceItemsHandler) Handle(ctx context.Context, cmd *RestoreTrashedWorkspaceItems) error {
	// TODO: domain.Note and domain.Folder have no Restore() method.
	// Add Restore() to both domain models (clears trashedBy and trashedAt fields),
	// then call note.Restore() / folder.Restore() before Save.
	// Also need to handle cascade restore for items trashed with TrashedByParent.
	for range cmd.NoteIDs {
		// # FIX: implement after Restore() is added to domain.Note
	}
	for range cmd.FolderIDs {
		// # FIX: implement after Restore() is added to domain.Folder
	}
	return nil
}
