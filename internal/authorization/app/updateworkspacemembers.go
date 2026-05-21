package app

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type UpdateWorkspaceMembers struct {
	UserID      string
	WorkspaceID uuid.UUID
	Members     []WorkspaceMember
}

type UpdateWorkspaceMembersHandler struct {
	enforcer             *casbin.TransactionalEnforcer
	integrationPublisher IntegrationPublisher
}

func NewUpdateWorkspaceMembersHandler(enforcer *casbin.TransactionalEnforcer, publisher IntegrationPublisher) *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{
		enforcer:             enforcer,
		integrationPublisher: publisher,
	}
}

var ProvideUpdateWorkspaceMembersHandler = NewUpdateWorkspaceMembersHandler

type UpdateWorkspaceMembersCmd commonhandler.Cmd[UpdateWorkspaceMembers]

var _ UpdateWorkspaceMembersCmd = (*UpdateWorkspaceMembersHandler)(nil)

func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, params *UpdateWorkspaceMembers) error {
	editAllowed, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace",
		WorkspacePermissionEdit.String(),
	)
	if err != nil {
		return errs.NewCasbinEnforcerError(err)
	}
	if !editAllowed {
		return errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, WorkspacePermissionEdit.String())
	}

	var oldMembers []WorkspaceMember
	var eventsToPublish []IntegrationEvent

	return h.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return errs.NewCasbinInternalError(err)
		}
		currentRules, err := bufferedModel.GetFilteredPolicy("g", "g", 2, formatWorkspace(params.WorkspaceID))
		if err != nil {
			return errs.NewCasbinInternalError(err)
		}

		oldMembers = make([]WorkspaceMember, 0, len(currentRules))
		for _, rule := range currentRules {
			userID, err := userFromFormat(rule[0])
			if err != nil {
				return errs.NewCasbinInternalError(err)
			}
			role, err := parseRole(rule[1])
			if err != nil {
				return errs.NewCasbinInternalError(err)
			}
			oldMembers = append(oldMembers, WorkspaceMember{
				ID:   userID,
				Role: role,
			})
		}

		for _, rule := range currentRules {
			_, err := tx.RemoveGroupingPolicy(rule[0], rule[1], rule[2])
			if err != nil {
				return errs.NewCasbinInternalError(err)
			}
		}

		for _, member := range params.Members {
			_, err := tx.AddNamedGroupingPolicy("g", formatUser(member.ID), member.Role.String(), formatWorkspace(params.WorkspaceID))
			if err != nil {
				return errs.NewCasbinInternalError(err)
			}
		}

		eventsToPublish = compareAndGenerateEvents(params.WorkspaceID, oldMembers, params.Members)

		if len(eventsToPublish) > 0 {
			if err := h.integrationPublisher.Publish(ctx, eventsToPublish...); err != nil {
				return errs.NewPublishIntegrationEventsFailed(params.WorkspaceID, err)
			}
		}

		return nil
	})
}

func compareAndGenerateEvents(workspaceID uuid.UUID, oldMembers, newMembers []WorkspaceMember) []IntegrationEvent {
	var events []IntegrationEvent

	oldMemberMap := make(map[string]WorkspaceRole)
	for _, member := range oldMembers {
		oldMemberMap[member.ID] = member.Role
	}

	newMemberMap := make(map[string]WorkspaceRole)
	for _, member := range newMembers {
		newMemberMap[member.ID] = member.Role
	}

	for userID, newRole := range newMemberMap {
		if oldRole, exists := oldMemberMap[userID]; !exists {
			events = append(events, IntegrationEventWorkspaceMemberAdded{
				WorkspaceID: workspaceID,
				UserID:      userID,
				Role:        newRole,
			})
		} else if oldRole != newRole {
			events = append(events, IntegrationEventUserWorkspaceRoleUpdated{
				WorkspaceID: workspaceID,
				UserID:      userID,
				Role:        newRole,
			})
		}
	}

	for userID := range oldMemberMap {
		if _, exists := newMemberMap[userID]; !exists {
			events = append(events, IntegrationEventWorkspaceMemberRemoved{
				WorkspaceID: workspaceID,
				UserID:      userID,
			})
		}
	}

	return events
}
