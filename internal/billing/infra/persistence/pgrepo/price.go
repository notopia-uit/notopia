package pgrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type priceRepo struct{ q *pgsqlc.Queries }

func (r *priceRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Price, error) {
	row, err := r.q.GetPrice(ctx, pgsqlc.GetPriceParams{ID: id, ForUpdate: forUpdate})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NewNotFound(fmt.Sprintf("price %s not found", id))
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get price", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) GetByProviderPriceID(ctx context.Context, id string) (*domain.Price, error) {
	row, err := r.q.GetPriceByProviderPriceID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NewNotFound(fmt.Sprintf("provider price %s not found", id))
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get price by provider id", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) GetActive(ctx context.Context, plan domain.PlanTier, interval domain.Interval, currency string, forUpdate bool) (*domain.Price, error) {
	row, err := r.q.GetActivePrice(ctx, &pgsqlc.GetActivePriceParams{
		Plan: pgmapper.FromPlanTier(plan), BillingInterval: pgmapper.FromInterval(interval), Currency: currency, ForUpdate: forUpdate,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get active price", err)
	}
	return pgmapper.ToPrice(row), nil
}

func (r *priceRepo) ListActive(ctx context.Context) ([]*domain.Price, error) {
	rows, err := r.q.ListActivePrices(ctx)
	if err != nil {
		return nil, errs.NewInternal("failed to list active prices", err)
	}
	return pgmapper.ToPrices(rows), nil
}

func (r *priceRepo) List(ctx context.Context) ([]*domain.Price, error) {
	rows, err := r.q.ListPrices(ctx)
	if err != nil {
		return nil, errs.NewInternal("failed to list prices", err)
	}
	return pgmapper.ToPrices(rows), nil
}

func (r *priceRepo) Save(ctx context.Context, p *domain.Price) error {
	err := r.q.UpsertPrice(ctx, &pgsqlc.UpsertPriceParams{
		ID: p.ID(), Plan: pgmapper.FromPlanTier(p.Plan()), BillingInterval: pgmapper.FromInterval(p.Interval()),
		Amount: p.Amount().Amount, Currency: p.Amount().Currency, ProviderPriceID: p.ProviderPriceID(),
		LookupKey: p.ProviderLookupKey(), Active: p.Active(), CreatedAt: p.CreatedAt(), ArchivedAt: p.ArchivedAt(),
	})
	drain(ctx, p.PopEvents())
	if err != nil {
		return errs.NewInternal("failed to save price", err)
	}
	return nil
}
