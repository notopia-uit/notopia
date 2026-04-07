package pgrepo

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
	publisher     Publisher       // Required when in transaction (should be tx-backed)
	inTransaction bool            // Indicates if we're already in a transaction
}

type RunInTxFnparams struct {
	queries   *pgsqlc.Queries
	publisher Publisher
}

// TODO: Can we just, avoid doing this cerr? common error tan` du
func runInTx(
	ctx context.Context,
	params *runInTxParams,
	fn func(params *RunInTxFnparams) error,
) (cerr error) {
	if params.inTransaction {
		return fn(&RunInTxFnparams{
			queries:   params.queries,
			publisher: params.publisher,
		})
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

	fnParams := &RunInTxFnparams{
		queries:   params.queries,
		publisher: params.publisher,
	}

	if err := fn(fnParams); err != nil {
		cerr = err
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction", err)
	}

	return nil
}
