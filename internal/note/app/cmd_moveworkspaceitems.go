package app

import (
	"context"
	"fmt"
	"log/slog"

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
	authorizationSvc AuthorizationSvc
	uow              domain.UnitOfWork
}

func NewMoveWorkspaceItemsHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *MoveWorkspaceItemsHandler {
	return &MoveWorkspaceItemsHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
	}
}

var ProvideMoveWorkspaceItemsHandler = NewMoveWorkspaceItemsHandler

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) error {
	slog.DebugContext(
		ctx, "moving workspace items",
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.Int("note_count", len(cmd.NoteIDs)),
		slog.Int("folder_count", len(cmd.FolderIDs)),
		slog.String("destination_folder_id", cmd.DestinationFolderID.String()),
		slog.String("user_id", cmd.UserID),
	)
	slog.DebugContext(
		ctx, "checking permission",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.String("permission", "write"),
	)
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionWrite)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to move items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}
	slog.DebugContext(
		ctx, "permission granted",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.WorkspaceID.String()),
	)

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
			slog.DebugContext(ctx, "folders validated", slog.Int("folder_count", len(cmd.FolderIDs)))
		}

		{
			noteValid, err := noteRepo.AreAllInWorkspace(ctx, cmd.NoteIDs, cmd.WorkspaceID)
			if err != nil {
				return err
			}
			if !noteValid {
				return errs.NewNotesNotInWorkspace(cmd.WorkspaceID)
			}
			slog.DebugContext(ctx, "notes validated", slog.Int("note_count", len(cmd.NoteIDs)))
		}

		{
			destinationFolder, err := folderRepo.GetByID(ctx, cmd.DestinationFolderID, false)
			if err != nil {
				return err
			}

			if destinationFolder.WorkspaceID() != cmd.WorkspaceID {
				return errs.NewDestinationFolderNotInWorkspace(cmd.DestinationFolderID, cmd.WorkspaceID)
			}
			slog.DebugContext(ctx, "destination folder validated", slog.String("destination_folder_id", cmd.DestinationFolderID.String()))
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
			slog.DebugContext(ctx, "folder hierarchy validated")
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
			slog.DebugContext(ctx, "folders moved", slog.Int("folder_count", len(folders)))
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
			slog.DebugContext(ctx, "notes moved", slog.Int("note_count", len(notes)))
		}
		slog.InfoContext(ctx, "workspace items moved successfully", slog.String("workspace_id", cmd.WorkspaceID.String()), slog.String("destination_folder_id", cmd.DestinationFolderID.String()))
		return nil
	})
}
