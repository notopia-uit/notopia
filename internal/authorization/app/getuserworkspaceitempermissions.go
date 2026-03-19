package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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
		return nil, newErrGetUserWorkspaceItemPermissionsCheckFailed(params.UserID, params.WorkspaceID)
	}
	wip := &WorkspaceItemPermissions{
		Read:   oks[0],
		Write:  oks[1],
		Delete: oks[2],
	}
	return wip, nil
}

var ErrCodeGetUserWorkspaceItemPermissionsCheckFailed = "GetUserWorkspaceItemPermissions_1"

func newErrGetUserWorkspaceItemPermissionsCheckFailed(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to check workspace item permissions for user %q on workspace %q", userID, workspaceID.String()),
		ErrCodeGetUserWorkspaceItemPermissionsCheckFailed,
		nil,
	)
}
