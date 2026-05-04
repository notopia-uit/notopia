package app

import (
	"context"
	"fmt"
	"log/slog"

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
	uow              domain.UnitOfWork
	trashService     *domain.TrashService
}

func NewTrashWorkspaceItemsHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
	trashService *domain.TrashService,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
		trashService:     trashService,
	}
}

var ProvideTrashWorkspaceItemsHandler = NewTrashWorkspaceItemsHandler

func (h *TrashWorkspaceItemsHandler) Handle(ctx context.Context, cmd *TrashWorkspaceItems) error {
	slog.DebugContext(
		ctx, "trashing workspace items",
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.Int("note_count", len(cmd.NoteIDs)),
		slog.Int("folder_count", len(cmd.FolderIDs)),
		slog.String("user_id", cmd.UserID),
	)
	slog.DebugContext(
		ctx, "checking permission",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.WorkspaceID.String()),
		slog.String("permission", "delete"),
	)
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
	slog.DebugContext(
		ctx, "permission granted",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.WorkspaceID.String()),
	)

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
			slog.DebugContext(ctx, "notes trashed", slog.Int("note_count", len(notes)))
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
			slog.DebugContext(ctx, "folders and children trashed", slog.Int("folder_count", len(folders)), slog.Int("child_folder_count", len(childFolders)), slog.Int("child_note_count", len(childNotes)))
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

		slog.InfoContext(ctx, "workspace items trashed successfully", slog.String("workspace_id", cmd.WorkspaceID.String()), slog.Int("total_notes", len(allModifiedNotes)), slog.Int("total_folders", len(allModifiedFolders)))
		return nil
	})
}
