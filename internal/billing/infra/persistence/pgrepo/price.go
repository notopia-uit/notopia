package pgrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type priceRepo struct{ q *pgsqlc.Queries }

func (r *priceRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Price, error) {
	row, err := r.q.GetPrice(ctx, pgsqlc.GetPriceParams{ID: id, ForUpdate: forUpdate})
	if notFound(err) {
		return nil, errs.NewNotFound(fmt.Sprintf("price %s not found", id))
	}
	if err != nil {
		return nil, wrap("get price", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) GetByProviderPriceID(ctx context.Context, id string) (*domain.Price, error) {
	row, err := r.q.GetPriceByProviderPriceID(ctx, id)
	if notFound(err) {
		return nil, errs.NewNotFound(fmt.Sprintf("provider price %s not found", id))
	}
	if err != nil {
		return nil, wrap("get price by provider id", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) GetActive(ctx context.Context, plan domain.PlanTier, interval domain.Interval, currency string, forUpdate bool) (*domain.Price, error) {
	row, err := r.q.GetActivePrice(ctx, &pgsqlc.GetActivePriceParams{
		Plan: pgmapper.FromPlanTier(plan), BillingInterval: pgmapper.FromInterval(interval), Currency: currency, ForUpdate: forUpdate,
	})
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, wrap("get active price", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) ListActive(ctx context.Context) ([]*domain.Price, error) {
	rows, err := r.q.ListActivePrices(ctx)
	return pgmapper.ToPrices(rows), wrap("list active prices", err)
}

func (r *priceRepo) List(ctx context.Context) ([]*domain.Price, error) {
	rows, err := r.q.ListPrices(ctx)
	return pgmapper.ToPrices(rows), wrap("list prices", err)
}

func (r *priceRepo) Save(ctx context.Context, p *domain.Price) error {
	err := r.q.UpsertPrice(ctx, &pgsqlc.UpsertPriceParams{
		ID: p.ID(), Plan: pgmapper.FromPlanTier(p.Plan()), BillingInterval: pgmapper.FromInterval(p.Interval()),
		Amount: p.Amount().Amount, Currency: p.Amount().Currency, ProviderPriceID: p.ProviderPriceID(),
		LookupKey: p.LookupKey(), Active: p.Active(), CreatedAt: p.CreatedAt(), ArchivedAt: p.ArchivedAt(),
	})
	drain(ctx, p.PopEvents())
	return wrap("save price", err)
}
