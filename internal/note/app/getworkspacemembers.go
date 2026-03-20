package app

import "context"

type GetWorkspaceMembers struct {
	Slug string
}

type GetWorkspaceMembersReadModel interface {
	GetWorkspaceMembers(ctx context.Context, q *GetWorkspaceMembers) (*[]WorkspaceMember, error)
}

type GetWorkspaceMembersHandler struct {
	readModel GetWorkspaceMembersReadModel
}

func NewGetWorkspaceMembersHandler(readModel GetWorkspaceMembersReadModel) *GetWorkspaceMembersHandler {
	return &GetWorkspaceMembersHandler{readModel: readModel}
}

var ProvideGetWorkspaceMembersHandler = NewGetWorkspaceMembersHandler

func (h *GetWorkspaceMembersHandler) Handle(ctx context.Context, query *GetWorkspaceMembers) (*[]WorkspaceMember, error) {
	return h.readModel.GetWorkspaceMembers(ctx, query)
}
