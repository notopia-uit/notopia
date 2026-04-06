package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type PermanentlyDeleteWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type PermanentlyDeleteWorkspaceItemsHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	folderRepo           domain.FolderRepo
}

func NewPermanentlyDeleteWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
) *PermanentlyDeleteWorkspaceItemsHandler {
	return &PermanentlyDeleteWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
	}
}

var ProvidePermanentlyDeleteWorkspaceItemsHandler = NewPermanentlyDeleteWorkspaceItemsHandler

func (h *PermanentlyDeleteWorkspaceItemsHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteWorkspaceItems) error {
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
			fmt.Sprintf("user %s does not have permission to permanently delete items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	if len(cmd.NoteIDs) > 0 {
		if err := h.noteRepo.PermanentlyDeleteByIDs(ctx, cmd.NoteIDs); err != nil {
			return err
		}
	}

	if len(cmd.FolderIDs) > 0 {
		if err := h.folderRepo.PermanentlyDeleteByIDs(ctx, cmd.FolderIDs); err != nil {
			return err
		}
	}

	return nil
}
