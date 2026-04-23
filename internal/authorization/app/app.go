package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v3"
)

type App struct {
	Enforcer                        *casbin.TransactionalEnforcer
	CreateWorkspace                 *CreateWorkspaceHandler
	DeleteWorkspace                 *DeleteWorkspaceHandler
	GetUserWorkspaceItemPermissions *GetUserWorkspaceItemPermissionsHandler
	GetUserWorkspaces               *GetUserWorkspacesHandler
	GetWorkspaceMembers             *GetWorkspaceMembersHandler
	HasWorkspaceItemPermission      *HasWorkspaceItemPermissionHandler
	HasWorkspacePermission          *HasWorkspacePermissionHandler
	UpdateWorkspaceMembers          *UpdateWorkspaceMembersHandler
}

func (a *App) BootStrapPolicies(ctx context.Context) error {
	slog.DebugContext(ctx, "BootStrapPolicies: adding permission policies")

	permissionPolicies := [][]string{
		// Owner permissions on workspace
		{"owner", "workspace", "read"},
		{"owner", "workspace", "edit"},
		{"owner", "workspace", "delete"},
		// Owner permissions on workspace_item
		{"owner", "workspace_item", "read"},
		{"owner", "workspace_item", "write"},
		{"owner", "workspace_item", "delete"},
		// Editor permissions
		{"editor", "workspace", "read"},
		{"editor", "workspace_item", "read"},
		{"editor", "workspace_item", "write"},
		{"editor", "workspace_item", "delete"},
		// Viewer permissions
		{"viewer", "workspace", "read"},
		{"viewer", "workspace_item", "read"},
	}

	_, err := a.Enforcer.AddPolicies(permissionPolicies)
	if err != nil {
		slog.ErrorContext(ctx, "BootStrapPolicies: failed to add permission policies", slog.Any("error", err))
		return fmt.Errorf("failed to add permission policies: %w", err)
	}

	roleInheritancePolicies := [][]string{
		// Role inheritance: note/folder inherit workspace_item (g2 type, 2 params)
		{"note", "workspace_item"},
		{"folder", "workspace_item"},
	}
	_, err = a.Enforcer.AddNamedGroupingPolicies("g2", roleInheritancePolicies)
	if err != nil {
		slog.ErrorContext(ctx, "BootStrapPolicies: failed to add role inheritance policies", slog.Any("error", err))
		return fmt.Errorf("failed to add role inheritance policies: %w", err)
	}

	slog.InfoContext(ctx, "BootStrapPolicies: permission policies added")
	return nil
}

func (a *App) SeedDev(ctx context.Context) error {
	slog.DebugContext(ctx, "SeedDev: seeding dev user workspace policies")

	devPolicies := [][]string{
		{"user:111", "owner", "workspace:00000000-0000-0000-0000-000000000111"},
		{"user:112", "editor", "workspace:00000000-0000-0000-0000-000000000111"},
		{"user:110", "viewer", "workspace:00000000-0000-0000-0000-000000000111"},
		{"user:112", "owner", "workspace:00000000-0000-0000-0000-000000000112"},
		{"user:111", "editor", "workspace:00000000-0000-0000-0000-000000000112"},
		{"user:110", "owner", "workspace:00000000-0000-0000-0000-000000000110"},
	}

	_, err := a.Enforcer.AddGroupingPolicies(devPolicies)
	if err != nil {
		return fmt.Errorf("failed to seed dev policies: %w", err)
	}

	slog.InfoContext(ctx, "SeedDev: dev user workspace policies seeded")
	return nil
}
