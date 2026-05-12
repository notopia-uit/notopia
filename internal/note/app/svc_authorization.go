package app

import (
	"context"

	"github.com/google/uuid"
)

type WorkspacePermission uint8

const (
	WorkspacePermissionUnspecified WorkspacePermission = iota
	WorkspacePermissionRead
	WorkspacePermissionEdit
	WorkspacePermissionDelete
)

func (p WorkspacePermission) String() string {
	switch p {
	case WorkspacePermissionRead:
		return "read"
	case WorkspacePermissionEdit:
		return "edit"
	case WorkspacePermissionDelete:
		return "delete"
	case WorkspacePermissionUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

type WorkspaceItemPermission uint8

const (
	WorkspaceItemPermissionUnspecified WorkspaceItemPermission = iota
	WorkspaceItemPermissionRead
	WorkspaceItemPermissionWrite
	WorkspaceItemPermissionDelete
)

func (p WorkspaceItemPermission) String() string {
	switch p {
	case WorkspaceItemPermissionRead:
		return "read"
	case WorkspaceItemPermissionWrite:
		return "write"
	case WorkspaceItemPermissionDelete:
		return "delete"
	case WorkspaceItemPermissionUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

type AuthorizationWorkspaceMember struct {
	ID   string
	Role WorkspaceRole
}

type AuthorizationUserWorkspace struct {
	ID   uuid.UUID // Workspace ID
	Role WorkspaceRole
}

type AuthorizationSvc interface {
	GetUserWorkspaces(
		ctx context.Context,
		userID string,
	) ([]AuthorizationUserWorkspace, error)

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

	CreateWorkspaceWithOwner(
		ctx context.Context,
		ownerID string,
		workspaceID uuid.UUID,
	) error

	UpdateWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		members []WorkspaceMemberUpdate,
	) error

	GetWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
	) ([]AuthorizationWorkspaceMember, error)

	DeleteWorkspace(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
	) error
}
