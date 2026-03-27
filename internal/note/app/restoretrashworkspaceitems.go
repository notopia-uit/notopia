package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RestoreTrashedWorkspaceItems struct {
	WorkspaceID uuid.UUID
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

func (h *RestoreTrashedWorkspaceItemsHandler) Handle(ctx context.Context, cmd *RestoreTrashedWorkspaceItems) errs.Error {
	trashedNotes, err := h.noteRepo.GetMany(ctx,
		//exhaustruct:ignore
		&domain.NoteRepoGetManyParams{
			WorkspaceID: &cmd.WorkspaceID,
			IsTrashed:   new(true),
		})
	if err != nil {
		return err
	}

	trashedFolders, err := h.folderRepo.GetMany(ctx,
		//exhaustruct:ignore
		&domain.FolderRepoGetManyParams{
			WorkspaceID: &cmd.WorkspaceID,
			IsTrashed:   new(true),
		})
	if err != nil {
		return err
	}

	trashedNotePtrs := trashedNotes
	trashedFolderPtrs := trashedFolders

	if len(cmd.NoteIDs) > 0 {
		notes, err := h.noteRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.NoteRepoGetManyParams{
				IDs:       cmd.NoteIDs,
				ForUpdate: true,
			})
		if err != nil {
			return err
		}

		notePtrs := notes
		for _, note := range notePtrs {
			if err := h.noteRepo.Save(ctx, note); err != nil {
				return err
			}
		}
	}

	if len(cmd.FolderIDs) > 0 {
		folders, err := h.folderRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.FolderRepoGetManyParams{
				IDs:       cmd.FolderIDs,
				ForUpdate: true,
			})
		if err != nil {
			return err
		}

		if err := h.trashService.RestoreFolders(&trashedNotePtrs, &trashedFolderPtrs, folders); err != nil {
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
