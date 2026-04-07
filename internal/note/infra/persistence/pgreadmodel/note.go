package pgreadmodel

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteReadModel = (*Note)(nil)

func GetNote(queries *pgsqlc.Queries) *Note {
	return &Note{queries: queries}
}

var ProvideNote = GetNote

func (h *Note) GetNote(ctx context.Context, q *app.GetNote) (*app.Note, error) {
	return nil, nil
}
