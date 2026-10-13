package pgrepo

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgjsonb"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type planRepo struct{ q *pgsqlc.Queries }

func (r *planRepo) Get(ctx context.Context, tier domain.PlanTier, forUpdate bool) (*domain.Plan, error) {
	row, err := r.q.GetPlan(ctx, pgsqlc.GetPlanParams{Tier: pgmapper.FromPlanTier(tier), ForUpdate: forUpdate})
	if notFound(err) {
		return nil, errs.NewNotFound(fmt.Sprintf("plan %s not found", tier))
	}
	if err != nil {
		return nil, wrap("get plan", err)
	}
	return pgmapper.ToPlan(row), nil
}

func (r *planRepo) ListMap(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error) {
	rows, err := r.q.ListPlans(ctx)
	if err != nil {
		return nil, wrap("list plans", err)
	}
	return pgmapper.ToPlanMap(rows), nil
}

func (r *planRepo) Save(ctx context.Context, p *domain.Plan) error {
	err := r.q.UpsertPlan(ctx, &pgsqlc.UpsertPlanParams{
		Tier: pgmapper.FromPlanTier(p.Tier()), Features: pgjsonb.Features(p.Features()),
		Limits: pgjsonb.Limits(p.Limits()), UpdatedAt: p.UpdatedAt(),
	})
	drain(ctx, p.PopEvents())
	return wrap("save plan", err)
}
