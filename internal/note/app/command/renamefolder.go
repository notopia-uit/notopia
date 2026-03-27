package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameFolder struct {
	ID     uuid.UUID
	Name   string
	UserID string
}

type RenameFolderHandler struct {
	authorization service.Authorization
	folderRepo    domain.FolderRepo
}

func NewRenameFolderHandler(
	authorization service.Authorization,
	folderRepo domain.FolderRepo,
) *RenameFolderHandler {
	return &RenameFolderHandler{
		authorization: authorization,
		folderRepo:    folderRepo,
	}
}

var ProvideRenameFolderHandler = NewRenameFolderHandler

func (h *RenameFolderHandler) Handle(ctx context.Context, cmd *RenameFolder) errs.Error {
	workspaceID, err := h.folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		workspaceID,
		service.WorkspaceItemPermissionWrite,
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
