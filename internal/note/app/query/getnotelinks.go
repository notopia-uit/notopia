package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetNoteLinks struct {
	ID            uuid.UUID
	OutgoingLinks bool
	Backlinks     bool
}

type GetNoteLinksReadModel interface {
	GetNoteLinks(ctx context.Context, q *GetNoteLinks) (*NoteLinkResult, errs.Error)
}

type GetNoteLinksHandler struct {
	readModel GetNoteLinksReadModel
}

func NewGetNoteLinksHandler(readModel GetNoteLinksReadModel) *GetNoteLinksHandler {
	return &GetNoteLinksHandler{readModel: readModel}
}

var ProvideGetNoteLinksHandler = NewGetNoteLinksHandler

func (h *GetNoteLinksHandler) Handle(ctx context.Context, q *GetNoteLinks) (*NoteLinkResult, errs.Error) {
	return h.readModel.GetNoteLinks(ctx, q)
}
