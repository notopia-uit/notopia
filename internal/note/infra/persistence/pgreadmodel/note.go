package pgreadmodel

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetNote struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteReadModel = (*GetNote)(nil)

func NewGetNote(queries *pgsqlc.Queries) *GetNote {
	return &GetNote{queries: queries}
}

var ProvideGetNote = NewGetNote

func (h *GetNote) GetNote(ctx context.Context, q *app.GetNote) (*app.Note, error) {
	return nil, nil
}
