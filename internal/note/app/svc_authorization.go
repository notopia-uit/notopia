package app

import (
	"context"

	"github.com/google/uuid"
)

type WorkspacePermission string

const (
	WorkspacePermissionRead   WorkspacePermission = "read"
	WorkspacePermissionEdit   WorkspacePermission = "edit"
	WorkspacePermissionDelete WorkspacePermission = "delete"
)

func (p WorkspacePermission) String() string {
	return string(p)
}

type WorkspaceItemPermission string

const (
	WorkspaceItemPermissionRead   WorkspaceItemPermission = "read"
	WorkspaceItemPermissionWrite  WorkspaceItemPermission = "write"
	WorkspaceItemPermissionDelete WorkspaceItemPermission = "delete"
)

func (p WorkspaceItemPermission) String() string {
	return string(p)
}

type WorkspaceMemberInfo struct {
	ID   string
	Role WorkspaceRole
}

type AuthorizationService interface {
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
		ownerID string,
		workspaceID uuid.UUID,
	) error

	GetWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
	) ([]*WorkspaceMemberInfo, error)
}
