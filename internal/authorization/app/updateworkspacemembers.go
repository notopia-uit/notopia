package app

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
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

	editAllowed, err := hasWorkspacePermission(enforcer, params.UserID, params.WorkspaceID, WorkspacePermissionEdit)
	if err != nil {
		return err
	}
	if !editAllowed {
		return errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, WorkspacePermissionEdit.String())
	}

	err = enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return errs.NewCasbinInternalError(err)
		}
		currentRules, err := bufferedModel.GetFilteredPolicy("g", "g", 2, formatWorkspace(params.WorkspaceID))
		if err != nil {
			return errs.NewCasbinInternalError(err)
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

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
