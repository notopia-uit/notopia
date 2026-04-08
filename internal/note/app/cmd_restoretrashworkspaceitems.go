package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

// TODO: This should carefully recheck
// Transaction?

type RestoreTrashedWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type RestoreTrashedWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	folderRepo           domain.FolderRepo
	trashService         *domain.TrashService
	uow                  domain.UnitOfWork
}

func NewRestoreTrashedWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	trashService *domain.TrashService,
	uow domain.UnitOfWork,
) *RestoreTrashedWorkspaceItemsHandler {
	return &RestoreTrashedWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
		trashService:         trashService,
		uow:                  uow,
	}
}

var ProvideRestoreTrashedWorkspaceItemsHandler = NewRestoreTrashedWorkspaceItemsHandler

func (h *RestoreTrashedWorkspaceItemsHandler) Handle(ctx context.Context, cmd *RestoreTrashedWorkspaceItems) error {
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
			fmt.Sprintf("user %s does not have permission to restore items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		noteRepo := r.Note()
		folderRepo := r.Folder()

		trashedNotes, err := noteRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.NoteRepoGetManyParams{
				WorkspaceID: cmd.WorkspaceID,
				TrashOnly:   true,
			},
		)
		if err != nil {
			return err
		}

		trashedFolders, err := folderRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.FolderRepoGetManyParams{
				WorkspaceID: cmd.WorkspaceID,
				TrashOnly:   true,
			},
		)
		if err != nil {
			return err
		}

		trashedNotePtrs := trashedNotes
		trashedFolderPtrs := trashedFolders

		if len(cmd.NoteIDs) > 0 {
			notes, err := noteRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.NoteRepoGetManyParams{
					IDs:       cmd.NoteIDs,
					TrashOnly: true,
					ForUpdate: true,
				},
			)
			if err != nil {
				return err
			}

			notePtrs := notes
			if err := h.trashService.RestoreNotes(notePtrs, cmd.UserID); err != nil {
				return err
			}
			for _, note := range notePtrs {
				if err := noteRepo.Save(ctx, note); err != nil {
					return err
				}
			}
		}

		if len(cmd.FolderIDs) > 0 {
			folders, err := folderRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.FolderRepoGetManyParams{
					IDs:       cmd.FolderIDs,
					ForUpdate: true,
				},
			)
			if err != nil {
				return err
			}

			if err := h.trashService.RestoreFolders(&trashedNotePtrs, &trashedFolderPtrs, folders, cmd.UserID); err != nil {
				return err
			}

			for _, folder := range trashedFolderPtrs {
				if err := folderRepo.Save(ctx, folder); err != nil {
					return err
				}
			}

			for _, note := range trashedNotePtrs {
				if err := noteRepo.Save(ctx, note); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
