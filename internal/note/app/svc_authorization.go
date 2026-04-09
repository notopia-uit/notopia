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

	CreateWorkspaceWithOwner(
		ctx context.Context,
		ownerID string,
		workspaceID uuid.UUID,
	) error

	UpdateWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		members []WorkspaceMemberUpdate, // NOTE: Hey this struct is from the handler :v?
	) error

	GetWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
	) ([]*WorkspaceMemberInfo, error)
}
