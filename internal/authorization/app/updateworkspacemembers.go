package app

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type UpdateWorkspaceMembersHandler struct {
	Enforcer *casbin.TransactionalEnforcer
}

func (h *UpdateWorkspaceMembersHandler) Handle(
	ctx context.Context,
	userID string,
	workspaceID uuid.UUID,
	members []WorkspaceMember,
) error {
	enforcer := h.Enforcer

	if allowed, err := hasWorkspacePermission(enforcer, userID, workspaceID, WorkspacePermissionEdit); err != nil || !allowed {
		return newErrUpdateWorkspaceMembersNoPermission(userID, workspaceID)
	}

	err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return newErrUpdateWorkspaceMembersGetBufferedFail(workspaceID)
		}
		currentRules, err := bufferedModel.GetFilteredPolicy("g", "g", 2, formatWorkspace(workspaceID))
		if err != nil {
			return newErrUpdateWorkspaceMembersGetRulesFail(workspaceID)
		}

		for _, rule := range currentRules {
			_, err := tx.RemoveGroupingPolicy(rule[0], rule[1], rule[2])
			if err != nil {
				return newErrUpdateWorkspaceMembersRemoveFail(workspaceID)
			}
		}

		for _, member := range members {
			_, err := tx.AddNamedGroupingPolicy("g", formatUser(member.ID), member.Role.String(), formatWorkspace(workspaceID))
			if err != nil {
				return newErrUpdateWorkspaceMembersAddFail(workspaceID)
			}
		}

		return nil
	})
	if err != nil {
		return newErrUpdateWorkspaceMembersFail(workspaceID)
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
