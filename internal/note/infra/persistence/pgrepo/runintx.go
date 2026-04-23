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

type RunInTxFnParams struct {
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
	fn func(params *RunInTxFnParams) error,
) error {
	slog.DebugContext(ctx, "Starting RunInTx execution", slog.Bool("in_transaction", params.inTransaction))
	if params.inTransaction {
		return fn(&RunInTxFnParams{
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
		} else if err != nil {
			if err = tx.Rollback(ctx); err != nil {
				slog.ErrorContext(
					ctx, "failed to rollback transaction after error in RunInTx",
					slog.Any("error", err),
				)
			}
		}
	}()

	queries := pgsqlc.New(tx)
	publisher, err := r.publisherFactory.Create(ctx, tx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to create publisher in RunInTx", err)
	}

	fnParams := &RunInTxFnParams{
		queries:   queries,
		publisher: publisher,
	}

	if err := fn(fnParams); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction in RunInTx", err)
	}
	slog.DebugContext(ctx, "Completed RunInTx execution successfully")

	return nil
}
