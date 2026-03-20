package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type TrashWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type TrashWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	noterepo             domain.NoteRepo
	folderrepo           domain.FolderRepo
}

func NewTrashWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noterepo domain.NoteRepo,
	folderrepo domain.FolderRepo,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noterepo:             noterepo,
		folderrepo:           folderrepo,
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
		return newErrTrashWorkspaceItemsForbidden(cmd.UserID, cmd.WorkspaceID)
	}

	for _, noteID := range cmd.NoteIDs {
		note, err := h.noterepo.GetByID(ctx, noteID, true)
		if err != nil {
			return domain.NewErrNoteNotFound(noteID, err)
		}
		note.Trash(domain.TrashedByPurpose)
		if err := h.noterepo.Save(ctx, note); err != nil {
			return err
		}
	}

	for _, folderID := range cmd.FolderIDs {
		folder, err := h.folderrepo.GetByID(ctx, folderID, true)
		if err != nil {
			return domain.NewErrFolderNotFound(folderID, err)
		}
		folder.Trash(domain.TrashedByPurpose)
		if err := h.folderrepo.Save(ctx, folder); err != nil {
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

var ErrCodeTrashWorkspaceItemsForbidden = "TrashWorkspaceItems_1"

func newErrTrashWorkspaceItemsForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to trash items in workspace %q", userID, workspaceID.String()),
		ErrCodeTrashWorkspaceItemsForbidden,
		nil,
	)
}
