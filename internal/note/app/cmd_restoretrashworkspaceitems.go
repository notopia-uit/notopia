package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type RestoreTrashedWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type RestoreTrashedWorkspaceItemsHandler struct {
	noteRepo     domain.NoteRepo
	folderRepo   domain.FolderRepo
	trashService *domain.TrashService
}

func NewRestoreTrashedWorkspaceItemsHandler(
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	trashService *domain.TrashService,
) *RestoreTrashedWorkspaceItemsHandler {
	return &RestoreTrashedWorkspaceItemsHandler{
		noteRepo:     noteRepo,
		folderRepo:   folderRepo,
		trashService: trashService,
	}
}

var ProvideRestoreTrashedWorkspaceItemsHandler = NewRestoreTrashedWorkspaceItemsHandler

func (h *RestoreTrashedWorkspaceItemsHandler) Handle(ctx context.Context, cmd *RestoreTrashedWorkspaceItems) error {
	trashedNotes, err := h.noteRepo.GetMany(ctx,
		domain.NewNoteRepoGetManyParamsByWorkspaceID(cmd.WorkspaceID).
			WithIsTrashed(true),
	)
	if err != nil {
		return err
	}

	trashedFolders, err := h.folderRepo.GetMany(ctx,
		domain.NewFolderRepoGetManyParamsByWorkspaceID(cmd.WorkspaceID).
			WithTrashed(),
	)
	if err != nil {
		return err
	}

	trashedNotePtrs := trashedNotes
	trashedFolderPtrs := trashedFolders

	if len(cmd.NoteIDs) > 0 {
		notes, err := h.noteRepo.GetMany(ctx,
			domain.NewNoteRepoGetManyParamsByIDs(cmd.NoteIDs).
				WithForUpdate(),
		)
		if err != nil {
			return err
		}

		notePtrs := notes
		if err := h.trashService.RestoreNotes(notePtrs, cmd.UserID); err != nil {
			return err
		}
		for _, note := range notePtrs {
			if err := h.noteRepo.Save(ctx, note); err != nil {
				return err
			}
		}
	}

	if len(cmd.FolderIDs) > 0 {
		folders, err := h.folderRepo.GetMany(ctx,
			domain.NewFolderRepoGetManyParamsByIDs(cmd.FolderIDs).
				WithForUpdate(),
		)
		if err != nil {
			return err
		}

		if err := h.trashService.RestoreFolders(&trashedNotePtrs, &trashedFolderPtrs, folders, cmd.UserID); err != nil {
			return err
		}

		for _, folder := range trashedFolderPtrs {
			if err := h.folderRepo.Save(ctx, folder); err != nil {
				return err
			}
		}

		for _, note := range trashedNotePtrs {
			if err := h.noteRepo.Save(ctx, note); err != nil {
				return err
			}
		}
	}

	return nil
}
