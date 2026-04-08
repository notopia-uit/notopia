package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type TrashWorkspaceItems struct {
	UserID      string
	WorkspaceID uuid.UUID
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type TrashWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	uow                  domain.UnitOfWork
	trashService         *domain.TrashService
}

func NewTrashWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
	trashService *domain.TrashService,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorizationService: authorizationService,
		uow:                  uow,
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

	// TODO: Why it getting 4 times??
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		noteRepo := r.Note()
		folderRepo := r.Folder()

		if len(cmd.NoteIDs) == 0 && len(cmd.FolderIDs) == 0 {
			return nil
		}

		workspaceNotes, err := noteRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.NoteRepoGetManyParams{
				WorkspaceID: cmd.WorkspaceID,
			},
		)
		if err != nil {
			return err
		}

		workspaceFolders, err := folderRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.FolderRepoGetManyParams{
				WorkspaceID: cmd.WorkspaceID,
				ForUpdate:   true,
			},
		)
		if err != nil {
			return err
		}

		workspaceNotePtrs := workspaceNotes
		workspaceFolderPtrs := workspaceFolders

		var notes []*domain.Note
		if len(cmd.NoteIDs) > 0 {
			notes, err = noteRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.NoteRepoGetManyParams{
					IDs:       cmd.NoteIDs,
					ForUpdate: true,
				},
			)
			if err != nil {
				return err
			}

			if err := h.trashService.TrashNotes(notes, cmd.UserID); err != nil {
				return err
			}
		}

		var folders []*domain.Folder
		if len(cmd.FolderIDs) > 0 {
			folders, err = folderRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.FolderRepoGetManyParams{
					IDs:       cmd.FolderIDs,
					ForUpdate: true,
				},
			)
			if err != nil {
				return err
			}

			if err := h.trashService.TrashFolders(&workspaceNotePtrs, &workspaceFolderPtrs, folders, cmd.UserID); err != nil {
				return err
			}
		}

		// Save items
		if len(workspaceNotePtrs) > 0 {
			if err := noteRepo.SaveMany(ctx, workspaceNotePtrs); err != nil {
				return err
			}
		}

		if len(workspaceFolderPtrs) > 0 {
			if err := folderRepo.SaveMany(ctx, workspaceFolderPtrs); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
