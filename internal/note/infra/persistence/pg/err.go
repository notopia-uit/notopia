package pg

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

func toErr(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return errs.NewPersistenceInternal(
			"an unexpected error occurred, not a pg error",
			err,
		)
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.CheckViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.StringDataRightTruncationDataException,
		pgerrcode.InvalidTextRepresentation,
		pgerrcode.InvalidBinaryRepresentation,
		pgerrcode.SerializationFailure:
		return errs.NewPersistenceInvalid(
			"invalid data",
			err,
		)
	default:
		return errs.NewPersistenceInternal(
			"an unexpected error occurred",
			err,
		)
	}
}
