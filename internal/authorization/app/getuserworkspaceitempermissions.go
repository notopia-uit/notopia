package app

import (
	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type GetUserWorkspaceItemPermissions struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type GetUserWorkspaceItemPermissionsHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewGetUserWorkspaceItemPermissionsHandler(enforcer *casbin.TransactionalEnforcer) *GetUserWorkspaceItemPermissionsHandler {
	return &GetUserWorkspaceItemPermissionsHandler{enforcer: enforcer}
}

var ProvideGetUserWorkspaceItemPermissionsHandler = NewGetUserWorkspaceItemPermissionsHandler

func (h *GetUserWorkspaceItemPermissionsHandler) Handle(params GetUserWorkspaceItemPermissions) (*WorkspaceItemPermissions, error) {
	oks, err := h.enforcer.BatchEnforce(
		[][]any{
			{formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace_item", WorkspaceItemPermissionRead.String()},
			{formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace_item", WorkspaceItemPermissionWrite.String()},
			{formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace_item", WorkspaceItemPermissionDelete.String()},
		},
	)
	if err != nil {
		return nil, errs.NewCasbinEnforcerError(err)
	}
	wip := &WorkspaceItemPermissions{
		Read:   oks[0],
		Write:  oks[1],
		Delete: oks[2],
	}
	return wip, nil
}
