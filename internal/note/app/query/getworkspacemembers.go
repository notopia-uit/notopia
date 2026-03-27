package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceMembers struct {
	WorkspaceID uuid.UUID
}

type GetWorkspaceMembersHandler struct{}

func NewGetWorkspaceMembersHandler() *GetWorkspaceMembersHandler {
	return &GetWorkspaceMembersHandler{}
}

var ProvideGetWorkspaceMembersHandler = NewGetWorkspaceMembersHandler

func (h *GetWorkspaceMembersHandler) Handle(ctx context.Context, q *GetWorkspaceMembers) ([]*WorkspaceMember, errs.Error) {
	return nil, nil
}
