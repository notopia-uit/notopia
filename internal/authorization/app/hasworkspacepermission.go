package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type HasWorkspacePermissionHandler struct {
	Enforcer *casbin.TransactionalEnforcer
}

func (h *HasWorkspacePermissionHandler) Handle(
	userID string,
	workspaceID uuid.UUID,
	permission WorkspacePermission,
) (bool, error) {
	ok, err := h.Enforcer.Enforce(
		formatUser(userID),
		formatWorkspace(workspaceID),
		"workspace",
		permission.String(),
	)
	if err != nil {
		return false, newErrHasWorkspacePermissionCheckFailed(userID, workspaceID)
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
