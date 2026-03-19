package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type GetWorkspaceMembers struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type GetWorkspaceMembersHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewGetWorkspaceMembersHandler(enforcer *casbin.TransactionalEnforcer) *GetWorkspaceMembersHandler {
	return &GetWorkspaceMembersHandler{enforcer: enforcer}
}

var ProvideGetWorkspaceMembersHandler = NewGetWorkspaceMembersHandler

func (h *GetWorkspaceMembersHandler) Handle(params GetWorkspaceMembers) ([]WorkspaceMember, error) {
	viewAllowed, err := h.enforcer.Enforce(formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace", WorkspacePermissionRead.String())
	if err != nil {
		return nil, newErrGetWorkspaceMembersReadPermissionFailed(params.UserID, params.WorkspaceID)
	}
	if !viewAllowed {
		return nil, newErrGetWorkspaceMembersNoPermission(params.UserID, params.WorkspaceID)
	}

	rules, err := h.enforcer.GetFilteredGroupingPolicy(2, formatWorkspace(params.WorkspaceID))
	if err != nil {
		return nil, newErrGetWorkspaceMembersGetFailed(params.WorkspaceID)
	}

	members := make([]WorkspaceMember, 0, len(rules))
	for _, rule := range rules {
		if len(rule) != 3 {
			return nil, newErrGetWorkspaceMembersInvalidRule(params.WorkspaceID)
		}
		userID, err := userFromFormat(rule[0])
		if err != nil {
			return nil, newErrGetWorkspaceMembersInvalidUser(params.WorkspaceID)
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
