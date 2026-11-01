package pgrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type PolicyRepo struct {
	q         *pgsqlc.Queries
	publisher EventPublisher
	inTx      bool
}

var _ domain.BillingPolicyRepo = (*PolicyRepo)(nil)

func NewPolicyRepo(q *pgsqlc.Queries) *PolicyRepo {
	return &PolicyRepo{q: q, publisher: nil, inTx: false}
}

var ProvidePolicyRepo = NewPolicyRepo

func newPolicyRepoInTx(q *pgsqlc.Queries, publisher EventPublisher) *PolicyRepo {
	return &PolicyRepo{q: q, publisher: publisher, inTx: true}
}

func (r *PolicyRepo) Get(ctx context.Context, forUpdate bool) (*domain.BillingPolicy, error) {
	if err := requireLockTx(r.inTx, forUpdate); err != nil {
		return nil, err
	}
	row, err := r.q.GetPolicy(ctx, pgsqlc.GetPolicyParams{ForUpdate: forUpdate})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NewNotFound("billing policy not seeded")
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get policy", err)
	}
	return pgmapper.ToPolicy(row), nil
}

func (r *PolicyRepo) Save(ctx context.Context, p *domain.BillingPolicy) error {
	if err := requireTx(r.inTx, "save"); err != nil {
		return err
	}
	err := r.q.UpsertPolicy(ctx, &pgsqlc.UpsertPolicyParams{
		PastDueGraceDays:         int32(p.PastDueGraceDays()),
		LifetimeRefundWindowDays: int32(p.LifetimeRefundWindowDays()),
		PortalConfigurationID:    p.PortalConfigurationID(),
		UpdatedAt:                p.UpdatedAt(),
	})
	if err != nil {
		return errs.NewInternal("failed to save policy", err)
	}
	return r.publisher.Publish(ctx, p.PopEvents()...)
}
