package query

import "github.com/google/uuid"

type GetNote struct {
	ID uuid.UUID
}

type GetNoteReadModel interface {
	GetNote(*GetNote) (Note, error)
}

type GetNoteHandler struct {
	readModel GetNoteReadModel
}

func NewGetNoteHandler(readModel GetNoteReadModel) *GetNoteHandler {
	return &GetNoteHandler{readModel: readModel}
}

func (h *GetNoteHandler) Handle(query *GetNote) (Note, error) {
	return h.readModel.GetNote(query)
}
