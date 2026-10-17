package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type catalog struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func newCatalog(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *catalog {
	return &catalog{
		uow:      uow,
		provider: provider,
		settings: settings,
	}
}

// createPrice creates the provider price first (idempotent by our price id), then swaps the active row.
func (c *catalog) createPrice(
	ctx context.Context,
	id uuid.UUID,
	plan domain.PlanTier,
	interval domain.Interval,
	amount domain.Money,
	userID string,
) error {
	providerPriceID, err := c.provider.CreatePrice(ctx, PriceInput{
		PriceID:  id,
		Plan:     plan,
		Interval: interval,
		Amount:   amount,
	})
	if err != nil {
		return err
	}
	now := time.Now()
	var previous *domain.Price
	err = c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if previous, err = r.Price().GetActive(ctx, plan, interval, amount.Currency, true); err != nil {
			return err
		}
		if previous != nil {
			if err := previous.Archive(now, userID); err != nil {
				return err
			}
			if err := r.Price().Save(ctx, previous); err != nil {
				return err
			}
		}
		created, err := domain.NewPrice(id, plan, interval, amount, providerPriceID, now, userID)
		if err != nil {
			return err
		}
		return r.Price().Save(ctx, created)
	})
	if err != nil {
		return err
	}
	if previous != nil {
		if err := c.provider.ArchivePrice(ctx, previous.ProviderPriceID()); err != nil {
			return err
		}
	}
	return c.refreshPortal(ctx, userID)
}

// refreshPortal creates a new portal configuration for the current active prices.
func (c *catalog) refreshPortal(ctx context.Context, userID string) error {
	var prices []*domain.Price
	if err := c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		prices, err = r.Price().ListActive(ctx)
		return err
	}); err != nil {
		return err
	}
	id, err := c.provider.CreatePortalConfiguration(ctx, prices, c.settings.WebBaseURL)
	if err != nil {
		return err
	}
	return c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		policy, err := r.BillingPolicy().Get(ctx, true)
		if err != nil {
			return err
		}
		policy.SetPortalConfiguration(id, time.Now(), userID)
		return r.BillingPolicy().Save(ctx, policy)
	})
}
