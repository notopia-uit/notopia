package query

import "github.com/google/uuid"

type GetNotes struct {
	ID         uuid.UUID
	pagination *PaginationParams
}

type GetNotesReadModel interface {
	GetNotes(*GetNotes) (NotePaginated, error)
}

type GetNotesHandler struct {
	readModel GetNotesReadModel
}

func NewGetNotesHandler(readModel GetNotesReadModel) *GetNotesHandler {
	return &GetNotesHandler{readModel: readModel}
}

func (h *GetNotesHandler) Handle(query *GetNotes) (NotePaginated, error) {
	return h.readModel.GetNotes(query)
}
