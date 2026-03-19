package app

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type UpdateWorkspaceMembers struct {
	UserID      string
	WorkspaceID uuid.UUID
	Members     []WorkspaceMember
}

type UpdateWorkspaceMembersHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewUpdateWorkspaceMembersHandler(enforcer *casbin.TransactionalEnforcer) *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{enforcer: enforcer}
}

var ProvideUpdateWorkspaceMembersHandler = NewUpdateWorkspaceMembersHandler

func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, params UpdateWorkspaceMembers) error {
	enforcer := h.enforcer

	if allowed, err := hasWorkspacePermission(enforcer, params.UserID, params.WorkspaceID, WorkspacePermissionEdit); err != nil || !allowed {
		return newErrUpdateWorkspaceMembersNoPermission(params.UserID, params.WorkspaceID)
	}

	err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return newErrUpdateWorkspaceMembersGetBufferedFail(params.WorkspaceID)
		}
		currentRules, err := bufferedModel.GetFilteredPolicy("g", "g", 2, formatWorkspace(params.WorkspaceID))
		if err != nil {
			return newErrUpdateWorkspaceMembersGetRulesFail(params.WorkspaceID)
		}

		for _, rule := range currentRules {
			_, err := tx.RemoveGroupingPolicy(rule[0], rule[1], rule[2])
			if err != nil {
				return newErrUpdateWorkspaceMembersRemoveFail(params.WorkspaceID)
			}
		}

		for _, member := range params.Members {
			_, err := tx.AddNamedGroupingPolicy("g", formatUser(member.ID), member.Role.String(), formatWorkspace(params.WorkspaceID))
			if err != nil {
				return newErrUpdateWorkspaceMembersAddFail(params.WorkspaceID)
			}
		}

		return nil
	})
	if err != nil {
		return newErrUpdateWorkspaceMembersFail(params.WorkspaceID)
	}
	return nil
}

var (
	ErrCodeUpdateWorkspaceMembersNoPermission    = "UpdateWorkspaceMembers_1"
	ErrCodeUpdateWorkspaceMembersGetBufferedFail = "UpdateWorkspaceMembers_2"
	ErrCodeUpdateWorkspaceMembersGetRulesFail    = "UpdateWorkspaceMembers_3"
	ErrCodeUpdateWorkspaceMembersRemoveFail      = "UpdateWorkspaceMembers_4"
	ErrCodeUpdateWorkspaceMembersAddFail         = "UpdateWorkspaceMembers_5"
	ErrCodeUpdateWorkspaceMembersFail            = "UpdateWorkspaceMembers_6"
)

func newErrUpdateWorkspaceMembersNoPermission(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("User %q cannot edit workspace %q", userID, workspaceID.String()),
		ErrCodeUpdateWorkspaceMembersNoPermission,
		nil,
	)
}

func newErrUpdateWorkspaceMembersGetBufferedFail(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to get buffered model for workspace %q", workspaceID.String()),
		ErrCodeUpdateWorkspaceMembersGetBufferedFail,
		nil,
	)
}

func newErrUpdateWorkspaceMembersGetRulesFail(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to get current workspace members for workspace %q", workspaceID.String()),
		ErrCodeUpdateWorkspaceMembersGetRulesFail,
		nil,
	)
}

func newErrUpdateWorkspaceMembersRemoveFail(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to remove existing workspace member rule for workspace %q", workspaceID.String()),
		ErrCodeUpdateWorkspaceMembersRemoveFail,
		nil,
	)
}

func newErrUpdateWorkspaceMembersAddFail(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to add workspace member rule for workspace %q", workspaceID.String()),
		ErrCodeUpdateWorkspaceMembersAddFail,
		nil,
	)
}

func newErrUpdateWorkspaceMembersFail(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to update workspace members for workspace %q", workspaceID.String()),
		ErrCodeUpdateWorkspaceMembersFail,
		nil,
	)
}
