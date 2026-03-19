package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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

func (h *HasWorkspaceItemPermissionHandler) Handle(params HasWorkspaceItemPermission) (bool, error) {
	ok, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace_item",
		params.Permission.String(),
	)
	if err != nil {
		return false, newErrHasWorkspaceItemPermissionCheckFailed(params.UserID, params.WorkspaceID)
	}
	return ok, nil
}

var ErrCodeHasWorkspaceItemPermissionCheckFailed = "HasWorkspaceItemPermission_1"

func newErrHasWorkspaceItemPermissionCheckFailed(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to check workspace item permission for user %q on workspace %q", userID, workspaceID.String()),
		ErrCodeHasWorkspaceItemPermissionCheckFailed,
		nil,
	)
}
