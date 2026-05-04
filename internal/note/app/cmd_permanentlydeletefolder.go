package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type PermanentlyDeleteFolder struct {
	ID     uuid.UUID
	UserID string
}

type PermanentlyDeleteFolderHandler struct {
	authorizationSvc AuthorizationSvc
	uow              domain.UnitOfWork
}

func NewPermanentlyDeleteFolderHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *PermanentlyDeleteFolderHandler {
	return &PermanentlyDeleteFolderHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
	}
}

var ProvidePermanentlyDeleteFolderHandler = NewPermanentlyDeleteFolderHandler

// NOTE: We delegate the infra persistence to cascading delete things
// Fact, we should handle this in domain, not infra
func (h *PermanentlyDeleteFolderHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteFolder) error {
	slog.DebugContext(ctx, "permanently deleting folder", slog.String("folder_id", cmd.ID.String()), slog.String("user_id", cmd.UserID))
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		workspaceID, err := folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		slog.DebugContext(ctx, "checking permission", slog.String("user_id", cmd.UserID), slog.String("workspace_id", workspaceID.String()), slog.String("permission", "delete"))
		hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, workspaceID, WorkspaceItemPermissionDelete)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errs.NewForbidden(
				fmt.Sprintf("user %s does not have permission to delete folder %s", cmd.UserID, cmd.ID),
			)
		}
		slog.DebugContext(ctx, "permission granted", slog.String("user_id", cmd.UserID), slog.String("folder_id", cmd.ID.String()))
		folder, err := folderRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		folder.PermanentlyDelete(cmd.UserID)
		err = folderRepo.Save(ctx, folder)
		if err == nil {
			slog.InfoContext(ctx, "folder permanently deleted successfully", slog.String("folder_id", cmd.ID.String()))
		}
		return err
	})
}
