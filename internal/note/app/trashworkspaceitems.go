package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type TrashWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type TrashWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	folderRepo           domain.FolderRepo
	trashService         *domain.TrashService
}

func NewTrashWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	trashService *domain.TrashService,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
		trashService:         trashService,
	}
}

var ProvideTrashWorkspaceItemsHandler = NewTrashWorkspaceItemsHandler

func (h *TrashWorkspaceItemsHandler) Handle(ctx context.Context, cmd *TrashWorkspaceItems) error {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		WorkspaceItemPermissionDelete,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to trash items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	workspaceNotes, err := h.noteRepo.GetByWorkspaceID(ctx, domain.NoteRepoGetByWorkspaceIDParams{
		WorkspaceID: cmd.WorkspaceID,
		TrashedBy:   nil,
	})
	if err != nil {
		return err
	}

	workspaceFolders, err := h.folderRepo.GetByWorkspaceID(ctx, domain.FolderRepoGetByWorkspaceIDParams{
		WorkspaceID: cmd.WorkspaceID,
		TrashedBy:   nil,
	})
	if err != nil {
		return err
	}

	workspaceNotePtrs := make([]*domain.Note, len(workspaceNotes))
	for i := range workspaceNotes {
		workspaceNotePtrs[i] = &workspaceNotes[i]
	}

	workspaceFolderPtrs := make([]*domain.Folder, len(workspaceFolders))
	for i := range workspaceFolders {
		workspaceFolderPtrs[i] = &workspaceFolders[i]
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

		if err := h.trashService.TrashNotes(notePtrs); err != nil {
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

		if err := h.trashService.TrashFolders(&workspaceNotePtrs, &workspaceFolderPtrs, folderPtrs); err != nil {
			return err
		}

		for _, folder := range workspaceFolderPtrs {
			if err := h.folderRepo.Save(ctx, folder); err != nil {
				return err
			}
		}

		for _, note := range workspaceNotePtrs {
			if err := h.noteRepo.Save(ctx, note); err != nil {
				return err
			}
		}
	}

	return nil
}
