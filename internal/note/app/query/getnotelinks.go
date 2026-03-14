package query

import "github.com/google/uuid"

type GetNoteLinks struct {
	ID            uuid.UUID
	OutgoingLinks *bool
	Backlinks     *bool
}

type GetNoteLinksReadModel interface {
	GetNoteLinks(*GetNoteLinks) (NoteLinkResult, error)
}

type GetNoteLinksHandler struct {
	readModel GetNoteLinksReadModel
}

func NewGetNoteLinksHandler(readModel GetNoteLinksReadModel) *GetNoteLinksHandler {
	return &GetNoteLinksHandler{readModel: readModel}
}

func (h *GetNoteLinksHandler) Handle(query *GetNoteLinks) (NoteLinkResult, error) {
	return h.readModel.GetNoteLinks(query)
}
