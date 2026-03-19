package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type HasWorkspaceItemPermissionHandler struct {
	Enforcer *casbin.TransactionalEnforcer
}

func (h *HasWorkspaceItemPermissionHandler) Handle(
	userID string,
	workspaceID uuid.UUID,
	permission WorkspaceItemPermission,
) (bool, error) {
	ok, err := h.Enforcer.Enforce(
		formatUser(userID),
		formatWorkspace(workspaceID),
		"workspace_item",
		permission.String(),
	)
	if err != nil {
		return false, newErrHasWorkspaceItemPermissionCheckFailed(userID, workspaceID)
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
