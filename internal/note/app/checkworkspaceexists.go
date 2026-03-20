package app

import "context"

type CheckWorkspaceSlugExists struct {
	Slug string
}

type CheckWorkspaceSlugExistsReadModel interface {
	CheckWorkspaceSlugExists(ctx context.Context, q *CheckWorkspaceSlugExists) (*CheckWorkspaceSlugExistsResult, error)
}

type CheckWorkspaceSlugExistsHandler struct {
	readModel CheckWorkspaceSlugExistsReadModel
}

func NewCheckWorkspaceSlugExistsHandler(readModel CheckWorkspaceSlugExistsReadModel) *CheckWorkspaceSlugExistsHandler {
	return &CheckWorkspaceSlugExistsHandler{readModel: readModel}
}

var ProvideCheckWorkspaceSlugExistsHandler = NewCheckWorkspaceSlugExistsHandler

func (h *CheckWorkspaceSlugExistsHandler) Handle(ctx context.Context, query *CheckWorkspaceSlugExists) (*CheckWorkspaceSlugExistsResult, error) {
	return h.readModel.CheckWorkspaceSlugExists(ctx, query)
}
