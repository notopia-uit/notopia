package pg

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/notopia-uit/notopia/pkg/note/domain"
)

func toDomainError(err error) error {
	var pgErr *pgconn.PgError
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", domain.ErrNotFound, err)
	}
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.CheckViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.SerializationFailure:
		return fmt.Errorf("%w: %v", domain.ErrInvalid, err)
	case
		pgerrcode.StringDataRightTruncationDataException,
		pgerrcode.InvalidTextRepresentation,
		pgerrcode.InvalidBinaryRepresentation:
		return fmt.Errorf("%w: %v", domain.ErrInvalid, err)
	default:
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
}
