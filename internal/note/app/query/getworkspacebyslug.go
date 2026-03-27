package query

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceBySlug struct {
	Slug string
}

type GetWorkspaceBySlugReadModel interface {
	GetWorkspaceBySlug(ctx context.Context, q *GetWorkspaceBySlug) (*Workspace, errs.Error)
}

type GetWorkspaceHandler struct {
	readModel GetWorkspaceBySlugReadModel
}

func NewGetWorkspaceBySlugHandler(readModel GetWorkspaceBySlugReadModel) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{readModel: readModel}
}

var ProvideGetWorkspaceBySlugHandler = NewGetWorkspaceBySlugHandler

func (h *GetWorkspaceHandler) Handle(ctx context.Context, q *GetWorkspaceBySlug) (*Workspace, errs.Error) {
	return h.readModel.GetWorkspaceBySlug(ctx, q)
}
