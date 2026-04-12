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
	deleteAllowed, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace",
		WorkspacePermissionDelete.String(),
	)
	if err != nil {
		return errs.NewCasbinEnforcerError(err)
	}
	if !deleteAllowed {
		return errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, WorkspacePermissionDelete.String())
	}

	_, err = h.enforcer.RemoveFilteredGroupingPolicy(2, formatWorkspace(params.WorkspaceID))
	if err != nil {
		return errs.NewCasbinInternalError(err)
	}

	slog.InfoContext(ctx, "deleted workspace",
		slog.String("user_id", params.UserID),
		slog.String("workspace_id", params.WorkspaceID.String()),
	)
	return nil
}
