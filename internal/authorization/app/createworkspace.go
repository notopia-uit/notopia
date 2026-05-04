package app

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type CreateWorkspace struct {
	OwnerID     string
	WorkspaceID uuid.UUID
}

type CreateWorkspaceHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewCreateWorkspaceHandler(enforcer *casbin.TransactionalEnforcer) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{enforcer: enforcer}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, params CreateWorkspace) error {
	slog.DebugContext(ctx, "Handling create workspace", slog.String("owner_id", params.OwnerID), slog.String("workspace_id", params.WorkspaceID.String()))
	ok, err := h.enforcer.AddGroupingPolicy(
		formatUser(params.OwnerID),
		WorkspaceRoleOwner.String(),
		formatWorkspace(params.WorkspaceID),
	)
	if err != nil {
		return errs.NewCasbinInternalError(err)
	}
	if !ok {
		return errs.NewCreateWorkspaceExists(params.OwnerID, params.WorkspaceID)
	}
	slog.InfoContext(ctx, "created workspace",
		slog.String("owner_id", params.OwnerID),
		slog.String("workspace_id", params.WorkspaceID.String()),
	)
	return nil
}
