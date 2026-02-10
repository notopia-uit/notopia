package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/note/domain"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
}

var _ domain.NoteRepo = (*Note)(nil)

func (n *Note) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	result, err := n.queries.GetNote(ctx, id)
	if err != nil {
		return nil, toDomainError(err)
	}
	return noteToDomain(result), nil
}

func noteToDomain(note *pgsqlc.Note) *domain.Note {
	return domain.UnmarshalNote(
		note.ID,
		note.Title,
		note.DeletedAt,
	)
}
