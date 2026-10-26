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
	pool    *pgxpool.Pool
	queries *pgsqlc.Queries
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(pool *pgxpool.Pool, queries *pgsqlc.Queries) *UnitOfWork {
	return &UnitOfWork{pool: pool, queries: queries}
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
	if err := fn(&registry{q: u.queries.WithTx(tx)}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.NewInternal("failed to commit transaction", err)
	}
	return nil
}

type registry struct{ q *pgsqlc.Queries }

func (r *registry) Subscription() domain.SubscriptionRepo     { return &subscriptionRepo{q: r.q} }
func (r *registry) BillingAccount() domain.BillingAccountRepo { return &accountRepo{q: r.q} }
func (r *registry) Price() domain.PriceRepo                   { return &priceRepo{q: r.q} }
func (r *registry) Plan() domain.PlanRepo                     { return &planRepo{q: r.q} }
func (r *registry) BillingPolicy() domain.BillingPolicyRepo   { return &policyRepo{q: r.q} }
func (r *registry) ProviderEventLog() domain.ProviderEventLog { return &providerEventLog{q: r.q} }
