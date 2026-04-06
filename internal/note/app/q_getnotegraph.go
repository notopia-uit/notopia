package app

import (
	"context"
	"math"

	"github.com/google/uuid"
)

type GetNoteGraph struct {
	ID    uuid.UUID
	Depth int
}

type GetNoteGraphReadModel interface {
	GetNoteGraph(ctx context.Context, q *GetNoteGraph) (*Graph, error)
}

type GetNoteGraphHandler struct {
	readModel GetNoteGraphReadModel
}

func NewGetNoteGraphHandler(readModel GetNoteGraphReadModel) *GetNoteGraphHandler {
	return &GetNoteGraphHandler{readModel: readModel}
}

var ProvideGetNoteGraphHandler = NewGetNoteGraphHandler

func (h *GetNoteGraphHandler) Handle(ctx context.Context, query *GetNoteGraph) (*Graph, error) {
	// TODO: Auth
	if query.Depth <= 0 {
		query.Depth = math.MaxInt
	}
	return h.readModel.GetNoteGraph(ctx, query)
}
