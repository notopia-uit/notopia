package query

import "github.com/google/uuid"

type GetNotes struct {
	ID         uuid.UUID
	pagination *PaginationParams
}

type GetNotesReadModel interface {
	GetNotes(GetNotes) ([]Note, error)
}

type GetNotesHandler struct {
	readModel GetNotesReadModel
}

func NewGetNotesHandler(readModel GetNotesReadModel) *GetNotesHandler {
	return &GetNotesHandler{readModel: readModel}
}
