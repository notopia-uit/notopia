package query

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetNoteGraph struct {
	ID    uuid.UUID
	Depth int
}

type GetNoteGraphReadModel interface {
	GetNoteGraph(ctx context.Context, q *GetNoteGraph) (*Graph, errs.Error)
}

type GetNoteGraphHandler struct {
	readModel GetNoteGraphReadModel
}

func NewGetNoteGraphHandler(readModel GetNoteGraphReadModel) *GetNoteGraphHandler {
	return &GetNoteGraphHandler{readModel: readModel}
}

var ProvideGetNoteGraphHandler = NewGetNoteGraphHandler

func (h *GetNoteGraphHandler) Handle(ctx context.Context, q *GetNoteGraph) (*Graph, errs.Error) {
	if q.Depth <= 0 {
		q.Depth = math.MaxInt
	}
	return h.readModel.GetNoteGraph(ctx, q)
}
