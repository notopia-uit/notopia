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

type WorkspaceItemPermission string

var (
	WorkspaceItemPermissionRead   WorkspaceItemPermission = "read"
	WorkspaceItemPermissionWrite  WorkspaceItemPermission = "write"
	WorkspaceItemPermissionDelete WorkspaceItemPermission = "delete"
)

type WorkspaceRole string

var (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleEditor WorkspaceRole = "editor"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

type WorkspaceMember struct {
	ID   uuid.UUID
	Role WorkspaceRole
}

type Authorization interface {
	HasWorkspacePermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission WorkspacePermission) (bool, error)
	HasWorkspaceItemPermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission WorkspaceItemPermission) (bool, error)
	HasWorkspaceNotePermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission WorkspaceItemPermission) (bool, error)
	HasWorkspaceFolderPermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission WorkspaceItemPermission) (bool, error)
	CreateWorkspaceWithOwnership(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, ownerID uuid.UUID) error
	GetWorkspaceMembers(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]WorkspaceMember, error)
}
