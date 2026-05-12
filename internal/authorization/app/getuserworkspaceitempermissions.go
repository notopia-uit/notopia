package app

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type GetUserWorkspaceItemPermissions struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type GetUserWorkspaceItemPermissionsHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewGetUserWorkspaceItemPermissionsHandler(enforcer *casbin.TransactionalEnforcer) *GetUserWorkspaceItemPermissionsHandler {
	return &GetUserWorkspaceItemPermissionsHandler{enforcer: enforcer}
}

var ProvideGetUserWorkspaceItemPermissionsHandler = NewGetUserWorkspaceItemPermissionsHandler

func (h *GetUserWorkspaceItemPermissionsHandler) Handle(ctx context.Context, params *GetUserWorkspaceItemPermissions) (WorkspaceItemPermissions, error) {
	readAllowed, err := h.enforcer.Enforce(
		formatUser(params.UserID),
		formatWorkspace(params.WorkspaceID),
		"workspace",
		WorkspacePermissionRead.String(),
	)
	if err != nil {
		return WorkspaceItemPermissions{}, errs.NewCasbinEnforcerError(err)
	}
	if !readAllowed {
		return WorkspaceItemPermissions{}, errs.NewMemberHasNoPermission(params.UserID, params.WorkspaceID, WorkspacePermissionRead.String())
	}

	oks, err := h.enforcer.BatchEnforce(
		[][]any{
			{formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace_item", WorkspaceItemPermissionRead.String()},
			{formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace_item", WorkspaceItemPermissionWrite.String()},
			{formatUser(params.UserID), formatWorkspace(params.WorkspaceID), "workspace_item", WorkspaceItemPermissionDelete.String()},
		},
	)
	if err != nil {
		return WorkspaceItemPermissions{}, errs.NewCasbinEnforcerError(err)
	}
	wip := WorkspaceItemPermissions{
		Read:   oks[0],
		Write:  oks[1],
		Delete: oks[2],
	}
	slog.DebugContext(
		ctx, "retrieved user workspace item permissions",
		slog.String("user_id", params.UserID),
		slog.String("workspace_id", params.WorkspaceID.String()),
	)
	return wip, nil
}
