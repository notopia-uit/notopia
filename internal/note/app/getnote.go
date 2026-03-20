package app

import (
	"context"

	"github.com/google/uuid"
)

type GetNote struct {
	ID uuid.UUID
}

type GetNoteReadModel interface {
	GetNote(ctx context.Context, q *GetNote) (Note, error)
}

type GetNoteHandler struct {
	readModel GetNoteReadModel
}

func NewGetNoteHandler(readModel GetNoteReadModel) *GetNoteHandler {
	return &GetNoteHandler{readModel: readModel}
}

func (h *GetNoteHandler) Handle(ctx context.Context, query *GetNote) (Note, error) {
	return h.readModel.GetNote(ctx, query)
}
