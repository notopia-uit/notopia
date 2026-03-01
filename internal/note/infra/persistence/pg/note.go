package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
}

var _ domain.NoteRepo = (*Note)(nil)

func (n *Note) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	result, err := n.queries.GetNote(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNoteNotFound(id, err)
		}
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
