package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type GetUserWorkspaces struct {
	UserID string
}

type GetUserWorkspacesHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewGetUserWorkspacesHandler(enforcer *casbin.TransactionalEnforcer) *GetUserWorkspacesHandler {
	return &GetUserWorkspacesHandler{enforcer: enforcer}
}

var ProvideGetUserWorkspacesHandler = NewGetUserWorkspacesHandler

func (h *GetUserWorkspacesHandler) Handle(ctx context.Context, params *GetUserWorkspaces) ([]UserWorkspace, error) {
	rules, err := h.enforcer.GetFilteredGroupingPolicy(0, formatUser(params.UserID))
	if err != nil {
		return nil, errs.NewCasbinInternalError(err)
	}

	workspaces := make([]UserWorkspace, 0, len(rules))
	for _, rule := range rules {
		if len(rule) != 3 {
			return nil, errs.NewCasbinPolicySignatureInvalid(fmt.Sprintf("expected 3 elements in grouping policy rule, got %d", len(rule)))
		}
		// TODO: wtf is this? maybe another function to parse the workspace ID
		workspaceID, err := uuid.Parse(rule[2][len("workspace:"):])
		if err != nil {
			return nil, errs.NewInvalid(fmt.Sprintf("invalid workspace ID in grouping policy rule: %s", rule[2]))
		}
		workspaces = append(workspaces, UserWorkspace{
			ID:   workspaceID,
			Role: WorkspaceRole(rule[1]),
		})
	}
	slog.DebugContext(
		ctx, "retrieved user workspaces",
		slog.String("user_id", params.UserID),
		slog.Int("workspace_count", len(workspaces)),
	)
	return workspaces, nil
}
