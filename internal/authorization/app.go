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

type WorkspaceItemPermissions struct {
	Read   bool
	Write  bool
	Delete bool
}

func formatUser(
	id string,
) string {
	return fmt.Sprintf("user:%s", id)
}

func userFromFormat(
	formatted string,
) (string, error) {
	if len(formatted) < 6 || formatted[:5] != "user:" {
		return "", fmt.Errorf("invalid formatted user: %s", formatted)
	}
	return formatted[5:], nil
}

func formatWorkspace(
	id uuid.UUID,
) string {
	return fmt.Sprintf("workspace:%s", id.String())
}

type App struct {
	enforcer *casbin.TransactionalEnforcer
}

func (a *App) CreateWorkspace(
	userID string,
	workspaceID uuid.UUID,
) error {
	ok, err := a.enforcer.AddGroupingPolicy(
		formatUser(userID),
		"owner",
		formatWorkspace(workspaceID),
	)
	if err != nil {
		return fmt.Errorf("failed to create workspace %s for user %s: %w", workspaceID, userID, err)
	}
	if !ok {
		return fmt.Errorf("failed to create workspace %s for user %s: policy already exists", workspaceID, userID)
	}
	return nil
}

func (a *App) UpdateWorkspaceMembers(
	ctx context.Context,
	userID string,
	workspaceID uuid.UUID,
	members []WorkspaceMember,
) error {
	err := a.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return fmt.Errorf("failed to get buffered model: %w", err)
		}

		enforcer, err := casbin.NewEnforcer(bufferedModel, nil)
		if err != nil {
			return fmt.Errorf("failed to create enforcer with buffered model: %w", err)
		}
		allowedToEdit, err := enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "workspace", WorkspacePermissionEdit.String())
		if err != nil {
			return fmt.Errorf("failed to check edit permission for user %s on workspace %s: %w", userID, workspaceID, err)
		}
		if !allowedToEdit {
			return fmt.Errorf("user %s does not have permission to edit workspace %s", userID, workspaceID)
		}

		currentRules, err := bufferedModel.GetFilteredPolicy("g", "g", 2, formatWorkspace(workspaceID))
		if err != nil {
			return fmt.Errorf("failed to get current workspace members: %w", err)
		}

		for _, rule := range currentRules {
			_, err := tx.RemoveGroupingPolicy(rule)
			if err != nil {
				return fmt.Errorf("failed to remove existing workspace member rule %v: %w", rule, err)
			}
		}

		for _, member := range members {
			_, err := tx.AddNamedGroupingPolicy("g", formatUser(member.ID), member.Role.String(), formatWorkspace(workspaceID))
			if err != nil {
				return fmt.Errorf("failed to add workspace member rule %v: %w", member, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update workspace members for workspace %s: %w", workspaceID, err)
	}
	return nil
}

func (a *App) GetWorkspaceMembers(
	userID string,
	workspaceID uuid.UUID,
) ([]WorkspaceMember, error) {
	viewAllowed, err := a.enforcer.Enforce(formatUser(userID), formatWorkspace(workspaceID), "workspace", WorkspacePermissionRead.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check read permission for user %s on workspace %s: %w", userID, workspaceID, err)
	}
	if !viewAllowed {
		return nil, fmt.Errorf("user %s does not have permission to view workspace %s members", userID, workspaceID)
	}

	rules, err := a.enforcer.GetFilteredGroupingPolicy(2, formatWorkspace(workspaceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace members for workspace %s: %w", workspaceID, err)
	}

	members := make([]WorkspaceMember, 0, len(rules))
	for _, rule := range rules {
		if len(rule) > 3 {
			return nil, fmt.Errorf("invalid policy rule: %v", rule)
		}
		userID, err := userFromFormat(rule[0])
		if err != nil {
			return nil, fmt.Errorf("invalid user format in policy rule: %v: %w", rule, err)
		}
		members = append(members, WorkspaceMember{
			ID:   userID,
			Role: WorkspaceRole(rule[1]),
		})
	}
	return members, nil
}

func (a *App) HasWorkspacePermission(
	userID string,
	workspaceID uuid.UUID,
	permission WorkspacePermission,
) (bool, error) {
	ok, err := a.enforcer.Enforce(
		formatUser(userID),
		formatWorkspace(workspaceID),
		"workspace",
		permission,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace permission for user %s on workspace %s: %w", userID, workspaceID, err)
	}
	return ok, nil
}

func (a *App) HasWorkspaceItemPermission(
	userID string,
	workspaceID uuid.UUID,
	permission WorkspaceItemPermission,
) (bool, error) {
	ok, err := a.enforcer.Enforce(
		formatUser(userID),
		formatWorkspace(workspaceID),
		"workspace_item",
		permission,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace item permission for user %s on workspace %s: %w", userID, workspaceID, err)
	}
	return ok, nil
}

func (a *App) GetUserWorkspaceItemPermissions(
	userID string,
	workspaceID uuid.UUID,
) (*WorkspaceItemPermissions, error) {
	oks, err := a.enforcer.BatchEnforce(
		[][]any{
			{formatUser(userID), formatWorkspace(workspaceID), "workspace_item", WorkspaceItemPermissionRead.String()},
			{formatUser(userID), formatWorkspace(workspaceID), "workspace_item", WorkspaceItemPermissionWrite.String()},
			{formatUser(userID), formatWorkspace(workspaceID), "workspace_item", WorkspaceItemPermissionDelete.String()},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace item permissions for user %s on workspace %s: %w", userID, workspaceID, err)
	}
	wip := &WorkspaceItemPermissions{
		Read:   oks[0],
		Write:  oks[1],
		Delete: oks[2],
	}
	return wip, nil
}

func (a *App) BootStrapPolicies() {
}

func (a *App) SeedDev() {
}
