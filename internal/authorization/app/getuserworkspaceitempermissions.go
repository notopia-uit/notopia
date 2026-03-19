package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type GetUserWorkspaceItemPermissionsHandler struct {
	Enforcer *casbin.TransactionalEnforcer
}

func (h *GetUserWorkspaceItemPermissionsHandler) Handle(
	userID string,
	workspaceID uuid.UUID,
) (*WorkspaceItemPermissions, error) {
	oks, err := h.Enforcer.BatchEnforce(
		[][]any{
			{formatUser(userID), formatWorkspace(workspaceID), "workspace_item", WorkspaceItemPermissionRead.String()},
			{formatUser(userID), formatWorkspace(workspaceID), "workspace_item", WorkspaceItemPermissionWrite.String()},
			{formatUser(userID), formatWorkspace(workspaceID), "workspace_item", WorkspaceItemPermissionDelete.String()},
		},
	)
	if err != nil {
		return nil, newErrGetUserWorkspaceItemPermissionsCheckFailed(userID, workspaceID)
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
