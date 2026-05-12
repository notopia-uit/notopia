package app

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type LeaveWorkspace struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type LeaveWorkspaceHandler struct {
	enforcer             *casbin.TransactionalEnforcer
	integrationPublisher IntegrationPublisher
}

func NewLeaveWorkspaceHandler(enforcer *casbin.TransactionalEnforcer, publisher IntegrationPublisher) *LeaveWorkspaceHandler {
	return &LeaveWorkspaceHandler{
		enforcer:             enforcer,
		integrationPublisher: publisher,
	}
}

var ProvideLeaveWorkspaceHandler = NewLeaveWorkspaceHandler

func (h *LeaveWorkspaceHandler) Handle(ctx context.Context, params LeaveWorkspace) error {
	err := h.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return errs.NewCasbinInternalError(err)
		}

		currentRules, err := bufferedModel.GetFilteredPolicy("g", "g", 2, formatWorkspace(params.WorkspaceID))
		if err != nil {
			return errs.NewCasbinInternalError(err)
		}

		var userRole WorkspaceRole
		userRuleIndex := -1
		ownerCount := 0

		for i, rule := range currentRules {
			ruleUserID, err := userFromFormat(rule[0])
			if err != nil {
				return errs.NewCasbinInternalError(err)
			}

			role, err := parseRole(rule[1])
			if err != nil {
				return errs.NewCasbinInternalError(err)
			}

			if ruleUserID == params.UserID {
				userRole = role
				userRuleIndex = i
			}

			if role == WorkspaceRoleOwner {
				ownerCount++
			}
		}

		if userRuleIndex == -1 {
			return errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, "member")
		}

		if userRole == WorkspaceRoleOwner && ownerCount == 1 {
			return errs.NewUserIsOnlyOwner(params.UserID, params.WorkspaceID)
		}

		_, err = tx.RemoveGroupingPolicy(currentRules[userRuleIndex][0], currentRules[userRuleIndex][1], currentRules[userRuleIndex][2])
		if err != nil {
			return errs.NewCasbinInternalError(err)
		}

		event := IntegrationEventWorkspaceMemberRemoved{
			WorkspaceID: params.WorkspaceID,
			UserID:      params.UserID,
		}

		if err := h.integrationPublisher.Publish(ctx, event); err != nil {
			return errs.NewPublishIntegrationEventsFailed(params.WorkspaceID, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.InfoContext(
		ctx, "user left workspace",
		slog.String("user_id", params.UserID),
		slog.String("workspace_id", params.WorkspaceID.String()),
	)
	return nil
}
