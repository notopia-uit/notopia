package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceGraph struct {
	ID     uuid.UUID
	Orphan *bool
}

type GetWorkspaceGraphReadModel interface {
	GetWorkspaceGraph(ctx context.Context, q *GetWorkspaceGraph) (*Graph, errs.Error)
}

type GetWorkspaceGraphHandler struct {
	readModel GetWorkspaceGraphReadModel
}

func NewGetWorkspaceGraphHandler(readModel GetWorkspaceGraphReadModel) *GetWorkspaceGraphHandler {
	return &GetWorkspaceGraphHandler{readModel: readModel}
}

var ProvideGetWorkspaceGraphHandler = NewGetWorkspaceGraphHandler

func (h *GetWorkspaceGraphHandler) Handle(ctx context.Context, q *GetWorkspaceGraph) (*Graph, errs.Error) {
	return h.readModel.GetWorkspaceGraph(ctx, q)
}
