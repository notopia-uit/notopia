package app

import (
	"context"

	"github.com/google/uuid"
)

type WorkspaceMemberUpdate struct {
	ID   uuid.UUID
	Role WorkspaceRole
}

type UpdateWorkspaceMembers struct {
	WorkspaceSlug string
	Members       []WorkspaceMemberUpdate
}

type UpdateWorkspaceMembersHandler struct{}

func NewUpdateWorkspaceMembersHandler() *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{}
}

var ProvideUpdateWorkspaceMembersHandler = NewUpdateWorkspaceMembersHandler

func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, cmd *UpdateWorkspaceMembers) error {
	// WARN: Unimplemented stub - no domain model for WorkspaceMember exists.
	// TODO: There is no domain model for WorkspaceMember. Implement a WorkspaceMember
	// domain entity and corresponding repo (WorkspaceMemberRepo) to manage member roles.
	// The repo should support bulk upsert of member-role assignments for a workspace.
	// Steps:
	// 1. Create domain/workspacemember.go with WorkspaceMember entity
	// 2. Create domain/workspacememberrepo.go with interface (Save, GetByWorkspaceID, etc.)
	// 3. Implement infra/persistence/pg/workspacemember.go
	// 4. Add repo to dependency injection
	// 5. Implement this handler to call repo.SaveMany(ctx, members)
	// 6. Consider publishing WorkspaceMembersUpdatedEvent for real-time updates
	return nil
}
