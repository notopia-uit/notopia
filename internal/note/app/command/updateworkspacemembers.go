package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/query"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type WorkspaceMemberUpdate struct {
	ID   uuid.UUID
	Role query.WorkspaceRole
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

func (h *UpdateWorkspaceMembersHandler) Handle(ctx context.Context, cmd *UpdateWorkspaceMembers) errs.Error {
	return nil
}
