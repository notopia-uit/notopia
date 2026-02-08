package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}
	}
	return noteToDomain(result), nil
}

func (n *Note) toDomainError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	switch pgErr.Code {
	case pgerrcode.UniqueViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.CheckViolation,
		pgerrcode.InvalidTextRepresentation,
		pgerrcode.InvalidBinaryRepresentation,
		pgerrcode.ForeignKeyViolation:
		return fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	default:
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
}

func noteToDomain(note *pgsqlc.Note) *domain.Note {
	return domain.UnmarshalNote(
		note.ID,
		note.Title,
		note.CreatedAt,
		note.UpdatedAt,
		note.DeletedAt,
	)
}
