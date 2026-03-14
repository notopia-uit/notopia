package command

import (
	"context"

	"github.com/google/uuid"
)

type WorkspaceMemberUpdate struct {
	ID   uuid.UUID
	Role string
}

type UpdateWorkspaceMembers struct {
	WorkspaceSlug string
	Members       []WorkspaceMemberUpdate
}

type UpdateWorkspaceMembersHandler struct{}

func NewUpdateWorkspaceMembersHandler() *UpdateWorkspaceMembersHandler {
	return &UpdateWorkspaceMembersHandler{}
}

func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, cmd *UpdateWorkspaceMembers) error {
	// TODO: There is no domain model for WorkspaceMember. Implement a WorkspaceMember
	// domain entity and corresponding repo (WorkspaceMemberRepo) to manage member roles.
	// The repo should support bulk upsert of member-role assignments for a workspace.
	return nil
}
