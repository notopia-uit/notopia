package pgrepo

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type UnitOfWork struct {
	pool             *pgxpool.Pool
	queries          *pgsqlc.Queries
	publisherFactory PublisherFactory
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(pool *pgxpool.Pool, queries *pgsqlc.Queries, publisherFactory PublisherFactory) *UnitOfWork {
	return &UnitOfWork{pool: pool, queries: queries, publisherFactory: publisherFactory}
}

var ProvideUnitOfWork = NewUnitOfWork

func (u *UnitOfWork) Execute(ctx context.Context, fn func(r domain.RepoRegistry) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return errs.NewInternal("failed to begin transaction", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rollback transaction", slog.Any("error", err))
		}
	}()
	publisher, err := u.publisherFactory.Create(tx)
	if err != nil {
		return errs.NewInternal("failed to create event publisher", err)
	}
	if err := fn(&registry{q: u.queries.WithTx(tx), publisher: publisher}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.NewInternal("failed to commit transaction", err)
	}
	return nil
}

type registry struct {
	q         *pgsqlc.Queries
	publisher EventPublisher
}

func (r *registry) Subscription() domain.SubscriptionRepo {
	return newSubscriptionRepoInTx(r.q, r.publisher)
}

func (r *registry) BillingAccount() domain.BillingAccountRepo {
	return newAccountRepoInTx(r.q, r.publisher)
}

func (r *registry) Price() domain.PriceRepo {
	return newPriceRepoInTx(r.q, r.publisher)
}

func (r *registry) Plan() domain.PlanRepo {
	return newPlanRepoInTx(r.q, r.publisher)
}

func (r *registry) BillingPolicy() domain.BillingPolicyRepo {
	return newPolicyRepoInTx(r.q, r.publisher)
}

func (r *registry) ProviderEventLog() domain.ProviderEventLog {
	return newProviderEventLogInTx(r.q)
}
