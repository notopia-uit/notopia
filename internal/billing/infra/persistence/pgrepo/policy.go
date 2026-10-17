package pgrepo

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type policyRepo struct{ q *pgsqlc.Queries }

func (r *policyRepo) Get(ctx context.Context, forUpdate bool) (*domain.BillingPolicy, error) {
	row, err := r.q.GetPolicy(ctx, pgsqlc.GetPolicyParams{ForUpdate: forUpdate})
	if notFound(err) {
		return nil, errs.NewNotFound("billing policy not seeded")
	}
	if err != nil {
		return nil, wrap("get policy", err)
	}
	return pgmapper.ToPolicy(row), nil
}

func (r *policyRepo) Save(ctx context.Context, p *domain.BillingPolicy) error {
	err := r.q.UpsertPolicy(ctx, &pgsqlc.UpsertPolicyParams{
		PastDueGraceDays:         int32(p.PastDueGraceDays()),
		LifetimeRefundWindowDays: int32(p.LifetimeRefundWindowDays()),
		PortalConfigurationID:    p.PortalConfigurationID(),
		UpdatedAt:                p.UpdatedAt(),
	})
	drain(ctx, p.PopEvents())
	return wrap("save policy", err)
}
