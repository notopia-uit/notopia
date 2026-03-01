package pg

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

var (
	ErrCodePersistenceInternal = "Persistence_1"
	ErrCodePersistenceInvalid  = "Persistence_2"
)

func toDomainError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return commonerror.NewInternal(
			"An unexpected error occurred",
			ErrCodePersistenceInvalid,
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
		return commonerror.NewInvalid(
			"Data integrity violation",
			ErrCodePersistenceInvalid,
			pgErr,
		)
	default:
		return commonerror.NewInternal(
			"An unexpected error occurred",
			ErrCodePersistenceInternal,
			pgErr,
		)
	}
}
