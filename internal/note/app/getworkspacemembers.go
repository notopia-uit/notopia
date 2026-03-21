package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

// TODO: Kev
type GetWorkspaceMembers struct {
	ID uuid.UUID
}

type GetWorkspaceMembersHandler struct{}

func NewGetWorkspaceMembersHandler() *GetWorkspaceMembersHandler {
	return &GetWorkspaceMembersHandler{}
}

var ProvideGetWorkspaceMembersHandler = NewGetWorkspaceMembersHandler

func (h *GetWorkspaceMembersHandler) Handle(ctx context.Context, query *GetWorkspaceMembers) ([]WorkspaceMember, errs.Error) {
	return nil, nil
}
