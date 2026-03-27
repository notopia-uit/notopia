package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
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

func (h *GetWorkspaceMembersHandler) Handle(params GetWorkspaceMembers) ([]*WorkspaceMember, error) {
	readAllowed, err := h.enforcer.Enforce(formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace", WorkspacePermissionRead.String())
	if err != nil {
		return nil, errs.NewCasbinEnforcerError(err)
	}
	if !readAllowed {
		return nil, errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, WorkspacePermissionRead.String())
	}

	rules, err := h.enforcer.GetFilteredGroupingPolicy(2, formatWorkspace(params.WorkspaceID))
	if err != nil {
		return nil, errs.NewCasbinInternalError(err)
	}

	members := make([]*WorkspaceMember, 0, len(rules))
	for _, rule := range rules {
		if len(rule) != 3 {
			return nil, errs.NewCasbinPolicySignatureInvalid(fmt.Sprintf("expected 3 elements in grouping policy rule, got %d", len(rule)))
		}
		userID, err := userFromFormat(rule[0])
		if err != nil {
			return nil, errs.NewInvalidUserFormat(rule[0], err)
		}
		members = append(members, &WorkspaceMember{
			ID:   userID,
			Role: WorkspaceRole(rule[1]),
		})
	}
	return members, nil
}
