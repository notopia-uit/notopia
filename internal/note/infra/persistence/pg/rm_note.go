package pg

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetNoteReadModel struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteReadModel = (*GetNoteReadModel)(nil)

func NewGetNoteReadModel(queries *pgsqlc.Queries) *GetNoteReadModel {
	return &GetNoteReadModel{queries: queries}
}

var ProvideGetNoteReadModel = NewGetNoteReadModel

func (h *GetNoteReadModel) GetNote(ctx context.Context, q *app.GetNote) (*app.Note, error) {
	return nil, nil
}
