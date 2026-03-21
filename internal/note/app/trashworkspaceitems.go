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
}

func NewTrashWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
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

	for _, noteID := range cmd.NoteIDs {
		note, err := h.noteRepo.GetByID(ctx, noteID, true)
		if err != nil {
			return err
		}
		note.Trash(domain.TrashedByPurpose)
		if err := h.noteRepo.Save(ctx, note); err != nil {
			return err
		}
	}

	for _, folderID := range cmd.FolderIDs {
		folder, err := h.folderRepo.GetByID(ctx, folderID, true)
		if err != nil {
			return err
		}
		folder.Trash(domain.TrashedByPurpose)
		if err := h.folderRepo.Save(ctx, folder); err != nil {
			return err
		}
		// WARN: Incomplete cascade logic - only trashes direct items, not children.
		// TODO: Cascade trash child notes and folders with TrashedByParent.
		// Requires NoteRepo/FolderRepo methods to list items by parentFolderID.
		// Steps:
		// 1. Add FolderRepo.GetByParentID(ctx, parentFolderID) -> []Folder
		// 2. Add NoteRepo.GetByFolderID(ctx, folderID) -> []Note (may exist)
		// 3. Recursively trash all children with TrashedByParent, not TrashedByPurpose
		// 4. Preserve original TrashedAt timestamp in parent for restore ordering
	}

	return nil
}
