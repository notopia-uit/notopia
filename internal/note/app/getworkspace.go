package app

import "context"

type GetWorkspace struct {
	Slug string
}

type GetWorkspaceReadModel interface {
	GetWorkspace(ctx context.Context, q *GetWorkspace) (*Workspace, error)
}

type GetWorkspaceHandler struct {
	readModel GetWorkspaceReadModel
}

func NewGetWorkspaceHandler(readModel GetWorkspaceReadModel) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{readModel: readModel}
}

var ProvideGetWorkspaceHandler = NewGetWorkspaceHandler

func (h *GetWorkspaceHandler) Handle(ctx context.Context, query *GetWorkspace) (*Workspace, error) {
	return h.readModel.GetWorkspace(ctx, query)
}
