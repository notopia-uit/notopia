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
	trashedNotesPurpose, err := h.noteRepo.GetByWorkspaceID(ctx, domain.NoteRepoGetByWorkspaceIDParams{
		WorkspaceID: cmd.WorkspaceID,
		TrashedBy:   &domain.TrashedByPurpose,
	})
	if err != nil {
		return err
	}

	trashedNotesParent, err := h.noteRepo.GetByWorkspaceID(ctx, domain.NoteRepoGetByWorkspaceIDParams{
		WorkspaceID: cmd.WorkspaceID,
		TrashedBy:   &domain.TrashedByParent,
	})
	if err != nil {
		return err
	}

	trashedNotes := append(trashedNotesPurpose, trashedNotesParent...)

	trashedFoldersPurpose, err := h.folderRepo.GetByWorkspaceID(ctx, domain.FolderRepoGetByWorkspaceIDParams{
		WorkspaceID: cmd.WorkspaceID,
		TrashedBy:   &domain.TrashedByPurpose,
	})
	if err != nil {
		return err
	}

	trashedFoldersParent, err := h.folderRepo.GetByWorkspaceID(ctx, domain.FolderRepoGetByWorkspaceIDParams{
		WorkspaceID: cmd.WorkspaceID,
		TrashedBy:   &domain.TrashedByParent,
	})
	if err != nil {
		return err
	}

	trashedFolders := append(trashedFoldersPurpose, trashedFoldersParent...)

	trashedNotePtrs := make([]*domain.Note, len(trashedNotes))
	for i := range trashedNotes {
		trashedNotePtrs[i] = &trashedNotes[i]
	}

	trashedFolderPtrs := make([]*domain.Folder, len(trashedFolders))
	for i := range trashedFolders {
		trashedFolderPtrs[i] = &trashedFolders[i]
	}

	if len(cmd.NoteIDs) > 0 {
		notes, err := h.noteRepo.GetByIDs(ctx, cmd.NoteIDs, true)
		if err != nil {
			return err
		}

		notePtrs := make([]*domain.Note, len(notes))
		for i := range notes {
			notePtrs[i] = &notes[i]
		}

		if err := h.trashService.RestoreNotes(notePtrs); err != nil {
			return err
		}

		for _, note := range notePtrs {
			if err := h.noteRepo.Save(ctx, note); err != nil {
				return err
			}
		}
	}

	if len(cmd.FolderIDs) > 0 {
		folders, err := h.folderRepo.GetByIDs(ctx, cmd.FolderIDs, true)
		if err != nil {
			return err
		}

		folderPtrs := make([]*domain.Folder, len(folders))
		for i := range folders {
			folderPtrs[i] = &folders[i]
		}

		if err := h.trashService.RestoreFolders(&trashedNotePtrs, &trashedFolderPtrs, folderPtrs); err != nil {
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
