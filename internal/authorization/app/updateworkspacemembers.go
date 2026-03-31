package app

import (
	"context"
	"log/slog"

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

	err = h.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
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
	slog.InfoContext(ctx, "updated workspace members",
		slog.String("user_id", params.UserID),
		slog.String("workspace_id", params.WorkspaceID.String()),
		slog.Int("member_count", len(params.Members)),
	)
	return nil
}
