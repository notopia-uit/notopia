package app

import "context"

type GetWorkspaceTree struct {
	Slug string
}

type GetWorkspaceTreeReadModel interface {
	GetWorkspaceTree(ctx context.Context, q *GetWorkspaceTree) (WorkspaceTreeFolder, error)
}

type GetWorkspaceTreeHandler struct {
	readModel GetWorkspaceTreeReadModel
}

func NewGetWorkspaceTreeHandler(readModel GetWorkspaceTreeReadModel) *GetWorkspaceTreeHandler {
	return &GetWorkspaceTreeHandler{readModel: readModel}
}

var ProvideGetWorkspaceTreeHandler = NewGetWorkspaceTreeHandler

func (h *GetWorkspaceTreeHandler) Handle(ctx context.Context, query *GetWorkspaceTree) (WorkspaceTreeFolder, error) {
	return h.readModel.GetWorkspaceTree(ctx, query)
}
