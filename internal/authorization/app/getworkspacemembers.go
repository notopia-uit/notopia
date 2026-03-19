package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type GetWorkspaceMembersHandler struct {
	Enforcer *casbin.TransactionalEnforcer
}

func (h *GetWorkspaceMembersHandler) Handle(
	userID string,
	workspaceID uuid.UUID,
) ([]WorkspaceMember, error) {
	viewAllowed, err := h.Enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "workspace", WorkspacePermissionRead.String())
	if err != nil {
		return nil, newErrGetWorkspaceMembersReadPermissionFailed(userID, workspaceID)
	}
	if !viewAllowed {
		return nil, newErrGetWorkspaceMembersNoPermission(userID, workspaceID)
	}

	rules, err := h.Enforcer.GetFilteredGroupingPolicy(2, formatWorkspace(workspaceID))
	if err != nil {
		return nil, newErrGetWorkspaceMembersGetFailed(workspaceID)
	}

	members := make([]WorkspaceMember, 0, len(rules))
	for _, rule := range rules {
		if len(rule) != 3 {
			return nil, newErrGetWorkspaceMembersInvalidRule(workspaceID)
		}
		userID, err := userFromFormat(rule[0])
		if err != nil {
			return nil, newErrGetWorkspaceMembersInvalidUser(workspaceID)
		}
		members = append(members, WorkspaceMember{
			ID:   userID,
			Role: WorkspaceRole(rule[1]),
		})
	}
	return members, nil
}

var (
	ErrCodeGetWorkspaceMembersReadPermissionFailed = "GetWorkspaceMembers_1"
	ErrCodeGetWorkspaceMembersNoPermission         = "GetWorkspaceMembers_2"
	ErrCodeGetWorkspaceMembersGetFailed            = "GetWorkspaceMembers_3"
	ErrCodeGetWorkspaceMembersInvalidRule          = "GetWorkspaceMembers_4"
	ErrCodeGetWorkspaceMembersInvalidUser          = "GetWorkspaceMembers_5"
)

func newErrGetWorkspaceMembersReadPermissionFailed(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to check read permission for user %q on workspace %q", userID, workspaceID.String()),
		ErrCodeGetWorkspaceMembersReadPermissionFailed,
		nil,
	)
}

func newErrGetWorkspaceMembersNoPermission(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("User %q does not have permission to view workspace %q members", userID, workspaceID.String()),
		ErrCodeGetWorkspaceMembersNoPermission,
		nil,
	)
}

func newErrGetWorkspaceMembersGetFailed(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to get workspace members for workspace %q", workspaceID.String()),
		ErrCodeGetWorkspaceMembersGetFailed,
		nil,
	)
}

func newErrGetWorkspaceMembersInvalidRule(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Invalid policy rule found for workspace %q", workspaceID.String()),
		ErrCodeGetWorkspaceMembersInvalidRule,
		nil,
	)
}

func newErrGetWorkspaceMembersInvalidUser(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Invalid user format in policy rule for workspace %q", workspaceID.String()),
		ErrCodeGetWorkspaceMembersInvalidUser,
		nil,
	)
}
