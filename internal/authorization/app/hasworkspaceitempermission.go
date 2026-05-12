package app

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type HasWorkspaceItemPermission struct {
	UserID      string
	WorkspaceID uuid.UUID
	Permission  WorkspaceItemPermission
}

type HasWorkspaceItemPermissionHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewHasWorkspaceItemPermissionHandler(enforcer *casbin.TransactionalEnforcer) *HasWorkspaceItemPermissionHandler {
	return &HasWorkspaceItemPermissionHandler{enforcer: enforcer}
}

var ProvideHasWorkspaceItemPermissionHandler = NewHasWorkspaceItemPermissionHandler

func (h *HasWorkspaceItemPermissionHandler) Handle(ctx context.Context, params *HasWorkspaceItemPermission) (bool, error) {
	ok, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace_item",
		params.Permission.String(),
	)
	if err != nil {
		return false, errs.NewCasbinEnforcerError(err)
	}
	return ok, nil
}
