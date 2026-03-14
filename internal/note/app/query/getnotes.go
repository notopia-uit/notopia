package query

import (
	"context"
	"github.com/google/uuid"
)

type GetNotes struct {
	ID         uuid.UUID
	pagination *PaginationParams
}

type GetNotesReadModel interface {
	GetNotes(ctx context.Context, q *GetNotes) (NotePaginated, error)
}

type GetNotesHandler struct {
	readModel GetNotesReadModel
}

func NewGetNotesHandler(readModel GetNotesReadModel) *GetNotesHandler {
	return &GetNotesHandler{readModel: readModel}
}

func (h *GetNotesHandler) Handle(ctx context.Context, query *GetNotes) (NotePaginated, error) {
	return h.readModel.GetNotes(ctx, query)
}
