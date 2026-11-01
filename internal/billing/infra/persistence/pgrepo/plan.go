package pgrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgjsonb"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type PlanRepo struct {
	q    *pgsqlc.Queries
	inTx bool
}

var _ domain.PlanRepo = (*PlanRepo)(nil)

func NewPlanRepo(q *pgsqlc.Queries) *PlanRepo {
	return &PlanRepo{q: q, inTx: false}
}

var ProvidePlanRepo = NewPlanRepo

func newPlanRepoInTx(q *pgsqlc.Queries) *PlanRepo {
	return &PlanRepo{q: q, inTx: true}
}

func (r *PlanRepo) Get(ctx context.Context, tier domain.PlanTier, forUpdate bool) (*domain.Plan, error) {
	if err := requireLockTx(r.inTx, forUpdate); err != nil {
		return nil, err
	}
	row, err := r.q.GetPlan(ctx, pgsqlc.GetPlanParams{Tier: pgmapper.FromPlanTier(tier), ForUpdate: forUpdate})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NewNotFound(fmt.Sprintf("plan %s not found", tier))
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get plan", err)
	}
	return pgmapper.ToPlan(row), nil
}

func (r *PlanRepo) ListMap(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error) {
	rows, err := r.q.ListPlans(ctx)
	if err != nil {
		return nil, errs.NewInternal("failed to list plans", err)
	}
	return pgmapper.ToPlanMap(rows), nil
}

func (r *PlanRepo) Save(ctx context.Context, p *domain.Plan) error {
	if err := requireTx(r.inTx, "save"); err != nil {
		return err
	}
	err := r.q.UpsertPlan(ctx, &pgsqlc.UpsertPlanParams{
		Tier: pgmapper.FromPlanTier(p.Tier()), Features: pgjsonb.Features(p.Features()),
		Limits: pgjsonb.Limits(p.Limits()), UpdatedAt: p.UpdatedAt(),
	})
	drain(ctx, p.PopEvents())
	if err != nil {
		return errs.NewInternal("failed to save plan", err)
	}
	return nil
}
