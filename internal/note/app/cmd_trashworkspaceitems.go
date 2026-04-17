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
	authorizationSvc AuthorizationSvc
	uow                  domain.UnitOfWork
	trashService         *domain.TrashService
}

func NewTrashWorkspaceItemsHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
	trashService *domain.TrashService,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorizationSvc: authorizationSvc,
		uow:                  uow,
		trashService:         trashService,
	}
}

var ProvideTrashWorkspaceItemsHandler = NewTrashWorkspaceItemsHandler

func (h *TrashWorkspaceItemsHandler) Handle(ctx context.Context, cmd *TrashWorkspaceItems) error {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
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

	if len(cmd.NoteIDs) == 0 && len(cmd.FolderIDs) == 0 {
		return nil
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		noteRepo := r.Note()
		folderRepo := r.Folder()

		var allModifiedNotes []*domain.Note
		var allModifiedFolders []*domain.Folder

		if len(cmd.NoteIDs) > 0 {
			notes, err := noteRepo.GetMany(ctx,
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

			allModifiedNotes = append(allModifiedNotes, notes...)
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

			var childFolders []*domain.Folder
			var childNotes []*domain.Note

			for _, folder := range folders {
				children, err := folderRepo.GetRecursiveChildren(ctx, &domain.FolderRepoGetRecursiveChildrenParams{
					ID:          folder.ID(),
					IncludeRoot: false,
					ForUpdate:   true,
				})
				if err != nil {
					return err
				}

				notes, err := noteRepo.GetRecursiveChildrenFromFolder(ctx, folder.ID(), true)
				if err != nil {
					return err
				}

				childFolders = append(childFolders, children...)
				childNotes = append(childNotes, notes...)
			}

			if err := h.trashService.TrashFoldersWithChildren(folders, childFolders, childNotes, cmd.UserID); err != nil {
				return err
			}

			allModifiedFolders = append(allModifiedFolders, folders...)
			allModifiedFolders = append(allModifiedFolders, childFolders...)
			allModifiedNotes = append(allModifiedNotes, childNotes...)
		}

		allModifiedNotes = deduplicateNotes(allModifiedNotes)
		allModifiedFolders = deduplicateFolders(allModifiedFolders)

		if len(allModifiedNotes) > 0 {
			if err := noteRepo.SaveMany(ctx, allModifiedNotes); err != nil {
				return err
			}
		}

		if len(allModifiedFolders) > 0 {
			if err := folderRepo.SaveMany(ctx, allModifiedFolders); err != nil {
				return err
			}
		}

		return nil
	})
}
