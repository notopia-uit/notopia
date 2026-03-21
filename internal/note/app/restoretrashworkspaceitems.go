package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RestoreTrashedWorkspaceItems struct {
	WorkspaceSlug string
	NoteIDs       []uuid.UUID
	FolderIDs     []uuid.UUID
}

type RestoreTrashedWorkspaceItemsHandler struct {
	noteRepo   domain.NoteRepo
	folderRepo domain.FolderRepo
}

func NewRestoreTrashedWorkspaceItemsHandler(
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
) *RestoreTrashedWorkspaceItemsHandler {
	return &RestoreTrashedWorkspaceItemsHandler{
		noteRepo:   noteRepo,
		folderRepo: folderRepo,
	}
}

var ProvideRestoreTrashedWorkspaceItemsHandler = NewRestoreTrashedWorkspaceItemsHandler

func (h *RestoreTrashedWorkspaceItemsHandler) Handle(ctx context.Context, cmd *RestoreTrashedWorkspaceItems) errs.Error {
	// WARN: Handler is completely stubbed - has no implementation.
	// TODO: domain.Note and domain.Folder have no Restore() method.
	// Add Restore() to both domain models (clears trashedBy and trashedAt fields),
	// then call note.Restore() / folder.Restore() before Save.
	// Also need to handle cascade restore for items trashed with TrashedByParent.
	// Steps:
	// 1. Add Restore() method to domain.Note: func (n *Note) Restore() { n.trashed = nil }
	// 2. Add Restore() method to domain.Folder: func (f *Folder) Restore() { f.trashed = nil }
	// 3. For each noteID in cmd.NoteIDs:
	//    - Get the note (with trashed=true to include trashed notes)
	//    - Call note.Restore()
	//    - If trashed.By == TrashedByParent, restore was automatic (parent restored first)
	//    - Save the note
	// 4. For each folderID in cmd.FolderIDs:
	//    - Similar logic to notes
	//    - Also restore all child items that were trashed with TrashedByParent
	// 5. Publish RestoreEvent for each restored item
	// 6. Consider cascade restore: when parent restores, children should too
	for range cmd.NoteIDs {
		// # WARN: Implement after Restore() is added to domain.Note
	}
	for range cmd.FolderIDs {
		// # WARN: Implement after Restore() is added to domain.Folder
	}
	return nil
}
