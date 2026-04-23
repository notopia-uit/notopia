package app

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type DeleteWorkspace struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type DeleteWorkspaceHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewDeleteWorkspaceHandler(enforcer *casbin.TransactionalEnforcer) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{enforcer: enforcer}
}

var ProvideDeleteWorkspaceHandler = NewDeleteWorkspaceHandler

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, params DeleteWorkspace) error {
	slog.DebugContext(ctx, "Handling delete workspace", slog.String("user_id", params.UserID), slog.String("workspace_id", params.WorkspaceID.String()))
	deleteAllowed, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace",
		WorkspacePermissionDelete.String(),
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to enforce policy", slog.String("user_id", params.UserID), slog.String("workspace_id", params.WorkspaceID.String()), slog.Any("error", err))
		return errs.NewCasbinEnforcerError(err)
	}
	if !deleteAllowed {
		slog.WarnContext(ctx, "permission denied for delete workspace", slog.String("user_id", params.UserID), slog.String("workspace_id", params.WorkspaceID.String()))
		return errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, WorkspacePermissionDelete.String())
	}

	_, err = h.enforcer.RemoveFilteredGroupingPolicy(2, formatWorkspace(params.WorkspaceID))
	if err != nil {
		slog.ErrorContext(ctx, "failed to remove grouping policy", slog.String("workspace_id", params.WorkspaceID.String()), slog.Any("error", err))
		return errs.NewCasbinInternalError(err)
	}

	slog.InfoContext(ctx, "deleted workspace",
		slog.String("user_id", params.UserID),
		slog.String("workspace_id", params.WorkspaceID.String()),
	)
	return nil
}
