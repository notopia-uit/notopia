package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type DeleteWorkspace struct {
	ID     uuid.UUID
	UserID string
}

type DeleteWorkspaceHandler struct {
	authorizationSvc AuthorizationSvc
	uow              domain.UnitOfWork
}

func NewDeleteWorkspaceHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
	}
}

var ProvideDeleteWorkspaceHandler = NewDeleteWorkspaceHandler

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd *DeleteWorkspace) error {
	slog.DebugContext(
		ctx, "checking permission",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.ID.String()),
		slog.String("permission", "delete"),
	)
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, cmd.UserID, cmd.ID, WorkspacePermissionDelete)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to delete workspace %s", cmd.UserID, cmd.ID),
		)
	}
	slog.DebugContext(
		ctx, "permission granted",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.ID.String()),
	)

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		workspaceRepo := r.Workspace()
		workspace, err := workspaceRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		workspace.Delete(cmd.UserID)
		if err := workspaceRepo.Save(ctx, workspace); err != nil {
			return err
		}
		return h.authorizationSvc.DeleteWorkspace(ctx, cmd.UserID, cmd.ID)
	})
}
