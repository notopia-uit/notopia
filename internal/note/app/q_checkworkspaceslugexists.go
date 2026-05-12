package app

import (
	"context"
)

type CheckWorkspaceSlugExists struct {
	Slug string
}

type CheckWorkspaceSlugExistsHandler struct {
	readModel CheckWorkspaceSlugExistsReadModel
}

func NewCheckWorkspaceSlugExistsHandler(readModel CheckWorkspaceSlugExistsReadModel) *CheckWorkspaceSlugExistsHandler {
	return &CheckWorkspaceSlugExistsHandler{readModel: readModel}
}

var ProvideCheckWorkspaceSlugExistsHandler = NewCheckWorkspaceSlugExistsHandler

func (h *CheckWorkspaceSlugExistsHandler) Handle(ctx context.Context, query *CheckWorkspaceSlugExists) (bool, error) {
	return h.readModel.CheckWorkspaceSlugExists(ctx, query.Slug)
}
