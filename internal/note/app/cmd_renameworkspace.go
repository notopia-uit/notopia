package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameWorkspace struct {
	ID     uuid.UUID
	Name   string
	UserID string
}

type RenameWorkspaceHandler struct {
	authorizationSvc AuthorizationSvc
	uow              domain.UnitOfWork
}

func NewRenameWorkspaceHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *RenameWorkspaceHandler {
	return &RenameWorkspaceHandler{
		authorizationSvc: authorizationSvc,
		uow:              uow,
	}
}

var ProvideRenameWorkspaceHandler = NewRenameWorkspaceHandler

func (h *RenameWorkspaceHandler) Handle(ctx context.Context, cmd *RenameWorkspace) error {
	slog.DebugContext(
		ctx, "checking permission",
		slog.String("user_id", cmd.UserID),
		slog.String("workspace_id", cmd.ID.String()),
		slog.String("permission", "edit"),
	)
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, cmd.UserID, cmd.ID, WorkspacePermissionEdit)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to rename workspace %s", cmd.UserID, cmd.ID),
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
		workspace.Rename(cmd.Name, cmd.UserID)
		return workspaceRepo.Save(ctx, workspace)
	})
}
