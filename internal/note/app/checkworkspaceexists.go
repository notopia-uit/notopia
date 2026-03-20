package app

import "context"

type CheckWorkspaceExists struct {
	Slug string
}

type CheckWorkspaceExistsReadModel interface {
	CheckWorkspaceExists(ctx context.Context, q *CheckWorkspaceExists) (*CheckWorkspaceExistsResult, error)
}

type CheckWorkspaceExistsHandler struct {
	readModel CheckWorkspaceExistsReadModel
}

func NewCheckWorkspaceExistsHandler(readModel CheckWorkspaceExistsReadModel) *CheckWorkspaceExistsHandler {
	return &CheckWorkspaceExistsHandler{readModel: readModel}
}

var ProvideCheckWorkspaceExistsHandler = NewCheckWorkspaceExistsHandler

func (h *CheckWorkspaceExistsHandler) Handle(ctx context.Context, query *CheckWorkspaceExists) (*CheckWorkspaceExistsResult, error) {
	return h.readModel.CheckWorkspaceExists(ctx, query)
}
