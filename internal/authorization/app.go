package authorization

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
)

type WorkspacePermission string

var (
	WorkspacePermissionRead   WorkspacePermission = "read"
	WorkspacePermissionEdit   WorkspacePermission = "edit"
	WorkspacePermissionDelete WorkspacePermission = "delete"
)

func (p WorkspacePermission) String() string {
	return string(p)
}

type WorkspaceItemPermission string

var (
	WorkspaceItemPermissionRead   WorkspaceItemPermission = "read"
	WorkspaceItemPermissionWrite  WorkspaceItemPermission = "write"
	WorkspaceItemPermissionDelete WorkspaceItemPermission = "delete"
)

func (p WorkspaceItemPermission) String() string {
	return string(p)
}

type WorkspaceRole string

var (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleEditor WorkspaceRole = "editor"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

func (r WorkspaceRole) String() string {
	return string(r)
}

type WorkspaceMember struct {
	ID   string
	Role WorkspaceRole
}

func formatUser(
	id string,
) string {
	return fmt.Sprintf("user:%s", id)
}

func formatWorkspace(
	id uuid.UUID,
) string {
	return fmt.Sprintf("workspace:%s", id.String())
}

type App struct {
	enforcer casbin.IEnforcerContext
}

func (a *App) BootStrapPolicies() {
}

func (a *App) SeedDev() {
}

func (a *App) HasWorkspacePermission(
	ctx context.Context,
	userID string,
	workspaceID uuid.UUID,
	permission WorkspacePermission,
) (bool, error) {
	ok, err := a.enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "workspace", permission)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace permission: %w", err)
	}
	return ok, nil
}

func (a *App) HasWorkspaceItemPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission WorkspaceItemPermission) bool {
	return a.enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "workspace_item", permission)
}

func (a *App) HasWorkspaceNotePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission WorkspaceItemPermission) bool {
	return a.enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "note", permission)
}

func (a *App) HasWorkspaceFolderPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission WorkspaceItemPermission) bool {
	return a.enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "folder", permission)
}

func (a *App) CreateWorkspaceWithOwnership(ctx context.Context, userID string, workspaceID uuid.UUID, ownerID uuid.UUID) error {
	_, err := a.enforcer.AddGroupingPolicy(formatUser(ownerID.String()), formatWorkspace(workspaceID), "workspace", WorkspaceRoleOwner)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) GetWorkspaceMembers(ctx context.Context, userID string, workspaceID uuid.UUID) ([]WorkspaceMember, error) {
	lines := a.enforcer.GetFilteredGroupingPolicy(2, formatWorkspace(workspaceID))

	members := make([]WorkspaceMember, 0, len(lines))
	for _, line := range lines {
		// line format: [user:id, role_name, workspace:id]
		members = append(members, WorkspaceMember{
			ID:   line[0], // You might want to strip the "user:" prefix here
			Role: WorkspaceRole(line[1]),
		})
	}

	return members, nil
}
