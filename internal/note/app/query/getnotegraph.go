package query

import "github.com/google/uuid"

type GetNoteGraph struct {
	ID    uuid.UUID
	Depth *int
}

type GetNoteGraphReadModel interface {
	GetNoteGraph(*GetNoteGraph) (Graph, error)
}

type GetNoteGraphHandler struct {
	readModel GetNoteGraphReadModel
}

func NewGetNoteGraphHandler(readModel GetNoteGraphReadModel) *GetNoteGraphHandler {
	return &GetNoteGraphHandler{readModel: readModel}
}

func (h *GetNoteGraphHandler) Handle(query *GetNoteGraph) (Graph, error) {
	return h.readModel.GetNoteGraph(query)
}
