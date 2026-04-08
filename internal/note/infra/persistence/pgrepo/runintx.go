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

type RunInTx struct {
	publisherFactory PublisherFactory
}

func NewRunInTx(publisherFactory PublisherFactory) *RunInTx {
	return &RunInTx{
		publisherFactory: publisherFactory,
	}
}

var ProvideRunInTx = NewRunInTx

func (r *RunInTx) Execute(
	ctx context.Context,
	params *runInTxParams,
	fn func(params *RunInTxFnparams) error,
) error {
	if params.inTransaction {
		return fn(&RunInTxFnparams{
			queries:   params.queries,
			publisher: params.publisher,
		})
	}

	tx, err := params.pgxPool.Begin(ctx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to begin transaction in RunInTx", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if err = tx.Rollback(ctx); err != nil {
				slog.ErrorContext(
					ctx, "failed to rollback transaction after panic in RunInTx",
					slog.Any("error", err),
				)
			}
			panic(p)
		}
	}()

	queries := pgsqlc.New(tx)
	publisher, err := r.publisherFactory.Create(tx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to create publisher in RunInTx", err)
	}

	fnParams := &RunInTxFnparams{
		queries:   queries,
		publisher: publisher,
	}

	if err := fn(fnParams); err != nil {
		if err = tx.Rollback(ctx); err != nil {
			slog.ErrorContext(
				ctx, "failed to rollback transaction after error in fn inside RunInTx",
				slog.Any("error", err),
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction in RunInTx", err)
	}

	return nil
}
