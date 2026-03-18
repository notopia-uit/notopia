package service

import (
	"context"

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

type Authorization interface {
	HasWorkspacePermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspacePermission,
	) (bool, error)

	HasWorkspaceItemPermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspaceItemPermission,
	) (bool, error)

	HasWorkspaceNotePermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspaceItemPermission,
	) (bool, error)

	HasWorkspaceFolderPermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspaceItemPermission,
	) (bool, error)

	CreateWorkspaceWithOwnership(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		ownerID uuid.UUID,
	) error

	GetWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
	) ([]WorkspaceMember, error)
}
