package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type WorkspacePermission uint8

const (
	WorkspacePermissionUnspecified WorkspacePermission = iota
	WorkspacePermissionBillingRead
	WorkspacePermissionBillingManage
)

func (p WorkspacePermission) IsValid() bool {
	switch p {
	case WorkspacePermissionBillingRead, WorkspacePermissionBillingManage:
		return true
	case WorkspacePermissionUnspecified:
		return false
	}
	return false
}

func (p WorkspacePermission) String() string {
	switch p {
	case WorkspacePermissionUnspecified:
		return "unspecified"
	case WorkspacePermissionBillingRead:
		return "billing_read"
	case WorkspacePermissionBillingManage:
		return "billing_manage"
	default:
		return "unknown"
	}
}

type WorkspaceRole uint8

const (
	WorkspaceRoleNone WorkspaceRole = iota
	WorkspaceRoleOwner
	WorkspaceRoleEditor
	WorkspaceRoleViewer
)

func (r WorkspaceRole) IsValid() bool {
	switch r {
	case WorkspaceRoleNone, WorkspaceRoleOwner, WorkspaceRoleEditor, WorkspaceRoleViewer:
		return true
	default:
		return false
	}
}

func (r WorkspaceRole) String() string {
	switch r {
	case WorkspaceRoleNone:
		return "none"
	case WorkspaceRoleOwner:
		return "owner"
	case WorkspaceRoleEditor:
		return "editor"
	case WorkspaceRoleViewer:
		return "viewer"
	default:
		return "unknown"
	}
}

type AuthorizationSvc interface {
	HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission WorkspacePermission) (bool, error)
	// GetMemberRole returns WorkspaceRoleNone when memberID is not a member. callerID must be a member.
	GetMemberRole(ctx context.Context, callerID string, workspaceID uuid.UUID, memberID string) (WorkspaceRole, error)
}

func requirePermission(ctx context.Context, authz AuthorizationSvc, userID string, workspaceID uuid.UUID, p WorkspacePermission) error {
	ok, err := authz.HasWorkspacePermission(ctx, userID, workspaceID, p)
	if err != nil {
		return err
	}
	if !ok {
		return errs.NewForbidden(fmt.Sprintf("user %s lacks billing permission on workspace %s", userID, workspaceID))
	}
	return nil
}
