package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
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

type WorkspaceRole string

var (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleEditor WorkspaceRole = "editor"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

type WorkspaceMemberInfo struct {
	ID   string
	Role WorkspaceRole
}

type Authorization interface {
	HasWorkspacePermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspacePermission,
	) (bool, errs.Error)

	HasWorkspaceItemPermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspaceItemPermission,
	) (bool, errs.Error)

	HasWorkspaceNotePermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspaceItemPermission,
	) (bool, errs.Error)

	HasWorkspaceFolderPermission(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		permission WorkspaceItemPermission,
	) (bool, errs.Error)

	CreateWorkspaceWithOwnership(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
		ownerID uuid.UUID,
	) errs.Error

	GetWorkspaceMembers(
		ctx context.Context,
		userID string,
		workspaceID uuid.UUID,
	) ([]*WorkspaceMemberInfo, errs.Error)
}
