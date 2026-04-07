package app

import (
	"context"
)

type GetWorkspaceBySlug struct {
	Slug string
}

type WorkspaceBySlugReadModel interface {
	GetWorkspaceBySlug(ctx context.Context, q *GetWorkspaceBySlug) (*Workspace, error)
}

type GetWorkspaceHandler struct {
	readModel WorkspaceBySlugReadModel
}

func NewGetWorkspaceBySlugHandler(readModel WorkspaceBySlugReadModel) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{readModel: readModel}
}

var ProvideGetWorkspaceBySlugHandler = NewGetWorkspaceBySlugHandler

func (h *GetWorkspaceHandler) Handle(ctx context.Context, query *GetWorkspaceBySlug) (*Workspace, error) {
	// TODO: Authorize
	return h.readModel.GetWorkspaceBySlug(ctx, query)
}
