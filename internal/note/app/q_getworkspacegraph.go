package app

import (
	"context"

	"github.com/google/uuid"
)

type GetWorkspaceGraph struct {
	ID            uuid.UUID
	IgnoreOrphans bool
}

type GetWorkspaceGraphReadModel interface {
	GetWorkspaceGraph(ctx context.Context, q *GetWorkspaceGraph) (*Graph, error)
}

type GetWorkspaceGraphHandler struct {
	readModel GetWorkspaceGraphReadModel
}

func NewGetWorkspaceGraphHandler(readModel GetWorkspaceGraphReadModel) *GetWorkspaceGraphHandler {
	return &GetWorkspaceGraphHandler{readModel: readModel}
}

var ProvideGetWorkspaceGraphHandler = NewGetWorkspaceGraphHandler

func (h *GetWorkspaceGraphHandler) Handle(ctx context.Context, query *GetWorkspaceGraph) (*Graph, error) {
	// TODO: Authorize
	return h.readModel.GetWorkspaceGraph(ctx, query)
}
