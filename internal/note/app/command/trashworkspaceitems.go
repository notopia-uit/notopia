package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type TrashWorkspaceItems struct {
	WorkspaceSlug string
	NoteIDs       []uuid.UUID
	FolderIDs     []uuid.UUID
}

type TrashWorkspaceItemsHandler struct {
	noterepo   domain.NoteRepo
	folderrepo domain.FolderRepo
}

func NewTrashWorkspaceItemsHandler(
	noterepo domain.NoteRepo,
	folderrepo domain.FolderRepo,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		noterepo:   noterepo,
		folderrepo: folderrepo,
	}
}

func (h *TrashWorkspaceItemsHandler) Handle(ctx context.Context, cmd *TrashWorkspaceItems) error {
	for _, noteID := range cmd.NoteIDs {
		note, err := h.noterepo.GetByID(ctx, noteID, true)
		if err != nil {
			return domain.NewErrNoteNotFound(noteID, err)
		}
		note.Trash(domain.TrashedByPurpose)
		if err := h.noterepo.Save(ctx, note); err != nil {
			return err
		}
	}

	for _, folderID := range cmd.FolderIDs {
		folder, err := h.folderrepo.GetByID(ctx, folderID, true)
		if err != nil {
			return domain.NewErrFolderNotFound(folderID, err)
		}
		folder.Trash(domain.TrashedByPurpose)
		if err := h.folderrepo.Save(ctx, folder); err != nil {
			return err
		}
		// TODO: Cascade trash child notes and folders with TrashedByParent.
		// Requires NoteRepo/FolderRepo methods to list items by parentFolderID.
	}

	return nil
}
