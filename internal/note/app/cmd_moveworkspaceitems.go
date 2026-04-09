package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type MoveWorkspaceItems struct {
	UserID              string
	WorkspaceID         uuid.UUID
	NoteIDs             []uuid.UUID
	FolderIDs           []uuid.UUID
	DestinationFolderID uuid.UUID
}

type MoveWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	uow                  domain.UnitOfWork
}

func NewMoveWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	uow domain.UnitOfWork,
) *MoveWorkspaceItemsHandler {
	return &MoveWorkspaceItemsHandler{
		authorizationService: authorizationService,
		uow:                  uow,
	}
}

var ProvideMoveWorkspaceItemsHandler = NewMoveWorkspaceItemsHandler

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) error {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionWrite)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to move items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		noteRepo := r.Note()

		{
			folderValid, err := folderRepo.AreAllInWorkspace(ctx, cmd.FolderIDs, cmd.WorkspaceID)
			if err != nil {
				return err
			}
			if !folderValid {
				return errs.NewFoldersNotInWorkspace(cmd.WorkspaceID)
			}
		}

		{
			noteValid, err := noteRepo.AreAllInWorkspace(ctx, cmd.NoteIDs, cmd.WorkspaceID)
			if err != nil {
				return err
			}
			if !noteValid {
				return errs.NewNotesNotInWorkspace(cmd.WorkspaceID)
			}
		}

		{
			destinationFolder, err := folderRepo.GetByID(ctx, cmd.DestinationFolderID, false)
			if err != nil {
				return err
			}

			if destinationFolder.WorkspaceID() != cmd.WorkspaceID {
				return errs.NewDestinationFolderNotInWorkspace(cmd.DestinationFolderID, cmd.WorkspaceID)
			}
		}

		{
			parentIDs, err := folderRepo.GetParentIDs(ctx, cmd.DestinationFolderID, true)
			if err != nil {
				return err
			}
			parentIDsSet := make(map[uuid.UUID]struct{})
			for _, id := range parentIDs {
				parentIDsSet[id] = struct{}{}
			}
			for _, folderID := range cmd.FolderIDs {
				if _, exists := parentIDsSet[folderID]; exists {
					return errs.NewCannotMoveFolderToItOwnSubfolder(folderID, cmd.DestinationFolderID)
				}
			}
		}

		if len(cmd.FolderIDs) == 0 && len(cmd.NoteIDs) == 0 {
			return nil
		}

		if len(cmd.FolderIDs) > 0 {
			folders, err := folderRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.FolderRepoGetManyParams{
					WorkspaceID: cmd.WorkspaceID,
					IDs:         cmd.FolderIDs,
					ForUpdate:   true,
				})
			if err != nil {
				return err
			}
			for _, folder := range folders {
				folder.MoveToFolder(cmd.DestinationFolderID, cmd.UserID)
			}
			if err := folderRepo.SaveMany(ctx, folders); err != nil {
				return err
			}
		}
		if len(cmd.NoteIDs) > 0 {
			notes, err := noteRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.NoteRepoGetManyParams{
					IDs:         cmd.NoteIDs,
					WorkspaceID: cmd.WorkspaceID,
					ForUpdate:   true,
				},
			)
			if err != nil {
				return err
			}
			for _, note := range notes {
				note.MoveToFolder(cmd.DestinationFolderID, cmd.UserID)
			}
			if err := noteRepo.SaveMany(ctx, notes); err != nil {
				return err
			}
		}
		return nil
	})
}
