package app

import (
	"fmt"

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
