package pg

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type runInTxParams struct {
	pgxPool       *pgxpool.Pool   // Required when not in transaction
	queries       *pgsqlc.Queries // Required when in transaction (should be tx-backed)
	inTransaction bool            // Indicates if we're already in a transaction
}

func runInTx(
	ctx context.Context,
	params *runInTxParams,
	fn func(queries *pgsqlc.Queries) errs.Error,
) (cerr errs.Error) {
	if params.inTransaction {
		return fn(params.queries)
	}

	tx, err := params.pgxPool.Begin(ctx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to begin transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if err = tx.Rollback(ctx); err != nil {
				slog.ErrorContext(
					ctx, "failed to rollback transaction after panic",
					slog.Any("error", err),
				)
			}
			panic(p)
		} else if cerr != nil {
			if err = tx.Rollback(ctx); err != nil {
				slog.ErrorContext(
					ctx, "failed to rollback transaction after error",
					slog.Any("error", err),
				)
			}
		}
	}()

	if err := fn(params.queries.WithTx(tx)); err != nil {
		cerr = err
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction", err)
	}

	return nil
}
