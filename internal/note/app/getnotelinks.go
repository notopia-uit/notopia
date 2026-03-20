package app

import (
	"context"
	"github.com/google/uuid"
)

type GetNoteLinks struct {
	ID            uuid.UUID
	OutgoingLinks *bool
	Backlinks     *bool
}

type GetNoteLinksReadModel interface {
	GetNoteLinks(ctx context.Context, q *GetNoteLinks) (*NoteLinkResult, error)
}

type GetNoteLinksHandler struct {
	readModel GetNoteLinksReadModel
}

func NewGetNoteLinksHandler(readModel GetNoteLinksReadModel) *GetNoteLinksHandler {
	return &GetNoteLinksHandler{readModel: readModel}
}

var ProvideGetNoteLinksHandler = NewGetNoteLinksHandler

func (h *GetNoteLinksHandler) Handle(ctx context.Context, query *GetNoteLinks) (*NoteLinkResult, error) {
	return h.readModel.GetNoteLinks(ctx, query)
}
