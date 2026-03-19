package app

import (
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

type WorkspaceItemPermissions struct {
	Read   bool
	Write  bool
	Delete bool
}

func formatUser(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func userFromFormat(formatted string) (string, error) {
	if len(formatted) < 6 || formatted[:5] != "user:" {
		return "", fmt.Errorf("invalid formatted user: %s", formatted)
	}
	return formatted[5:], nil
}

func formatWorkspace(id uuid.UUID) string {
	return fmt.Sprintf("workspace:%s", id.String())
}

func hasWorkspacePermission(
	enforcer *casbin.TransactionalEnforcer,
	userID string,
	workspaceID uuid.UUID,
	permission WorkspacePermission,
) (bool, error) {
	ok, err := enforcer.Enforce(
		formatUser(userID),
		formatWorkspace(workspaceID),
		"workspace",
		permission.String(),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace permission for user %s on workspace %s: %w", userID, workspaceID, err)
	}
	return ok, nil
}
