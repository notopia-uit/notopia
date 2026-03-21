package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameFolder struct {
	ID     uuid.UUID
	Name   string
	UserID string
}

type RenameFolderHandler struct {
	authorizationService AuthorizationService
	folderRepo           domain.FolderRepo
}

func NewRenameFolderHandler(
	authorizationService AuthorizationService,
	folderRepo domain.FolderRepo,
) *RenameFolderHandler {
	return &RenameFolderHandler{
		authorizationService: authorizationService,
		folderRepo:           folderRepo,
	}
}

var ProvideRenameFolderHandler = NewRenameFolderHandler

func (h *RenameFolderHandler) Handle(ctx context.Context, cmd *RenameFolder) errs.Error {
	workspaceID, err := h.folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to rename folder %s", cmd.UserID, cmd.ID),
		)
	}

	folder, err := h.folderRepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	folder.Rename(cmd.Name)
	return h.folderRepo.Save(ctx, folder)
}
