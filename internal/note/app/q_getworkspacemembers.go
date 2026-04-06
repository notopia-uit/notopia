package app

import (
	"context"

	"github.com/google/uuid"
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

func (h *GetWorkspaceMembersHandler) Handle(ctx context.Context, query *GetWorkspaceMembers) ([]*WorkspaceMember, error) {
	return nil, nil
}
