package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type CheckWorkspaceSlugExists struct {
	Slug string
}

type CheckWorkspaceSlugExistsReadModel interface {
	CheckWorkspaceSlugExists(ctx context.Context, q *CheckWorkspaceSlugExists) (*CheckWorkspaceSlugExistsResult, errs.Error)
}

type CheckWorkspaceSlugExistsHandler struct {
	readModel CheckWorkspaceSlugExistsReadModel
}

func NewCheckWorkspaceSlugExistsHandler(readModel CheckWorkspaceSlugExistsReadModel) *CheckWorkspaceSlugExistsHandler {
	return &CheckWorkspaceSlugExistsHandler{readModel: readModel}
}

var ProvideCheckWorkspaceSlugExistsHandler = NewCheckWorkspaceSlugExistsHandler

func (h *CheckWorkspaceSlugExistsHandler) Handle(ctx context.Context, query *CheckWorkspaceSlugExists) (*CheckWorkspaceSlugExistsResult, errs.Error) {
	return h.readModel.CheckWorkspaceSlugExists(ctx, query)
}
