package app

import (
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

func (h *HasWorkspacePermissionHandler) Handle(params HasWorkspacePermission) (bool, error) {
	ok, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace",
		params.Permission.String(),
	)
	if err != nil {
		return false, errs.NewCasbinEnforcerError(err)
	}
	return ok, nil
}
