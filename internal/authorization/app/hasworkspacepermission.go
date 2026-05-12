package app

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type HasWorkspacePermission struct {
	UserID      string
	WorkspaceID uuid.UUID
	Permission  WorkspacePermission
}

type HasWorkspacePermissionHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewHasWorkspacePermissionHandler(enforcer *casbin.TransactionalEnforcer) *HasWorkspacePermissionHandler {
	return &HasWorkspacePermissionHandler{enforcer: enforcer}
}

var ProvideHasWorkspacePermissionHandler = NewHasWorkspacePermissionHandler

func (h *HasWorkspacePermissionHandler) Handle(ctx context.Context, params *HasWorkspacePermission) (bool, error) {
	ok, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace",
		params.Permission.String(),
	)
	if err != nil {
		return false, errs.NewCasbinEnforcerError(err)
	}
	slog.DebugContext(
		ctx, "checked workspace permission",
		slog.String("user_id", params.UserID),
		slog.String("workspace_id", params.WorkspaceID.String()),
		slog.String("permission", params.Permission.String()),
		slog.Bool("allowed", ok),
	)
	return ok, nil
}
