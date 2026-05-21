package app

import (
	"context"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
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

type CheckWorkspaceSlugExistsQuery commonhandler.Query[CheckWorkspaceSlugExists, bool]

var _ CheckWorkspaceSlugExistsQuery = (*CheckWorkspaceSlugExistsHandler)(nil)

func (h *CheckWorkspaceSlugExistsHandler) Handle(ctx context.Context, query *CheckWorkspaceSlugExists) (bool, error) {
	return h.readModel.CheckWorkspaceSlugExists(ctx, query.Slug)
}
