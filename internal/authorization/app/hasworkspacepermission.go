package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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
		return false, newErrHasWorkspacePermissionCheckFailed(params.UserID, params.WorkspaceID)
	}
	return ok, nil
}

var ErrCodeHasWorkspacePermissionCheckFailed = "HasWorkspacePermission_1"

func newErrHasWorkspacePermissionCheckFailed(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to check workspace permission for user %q on workspace %q", userID, workspaceID.String()),
		ErrCodeHasWorkspacePermissionCheckFailed,
		nil,
	)
}
