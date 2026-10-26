package billingseed

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type defaultPrice struct {
	plan     app.PlanTier
	interval app.BillingInterval
	amount   int64
}

var defaultPrices = []defaultPrice{
	{plan: app.PlanTierPlus, interval: app.BillingIntervalMonth, amount: 49000},
	{plan: app.PlanTierPlus, interval: app.BillingIntervalYear, amount: 470400},
	{plan: app.PlanTierPro, interval: app.BillingIntervalMonth, amount: 99000},
	{plan: app.PlanTierPro, interval: app.BillingIntervalYear, amount: 950400},
	{plan: app.PlanTierPro, interval: app.BillingIntervalLifetime, amount: 2490000},
}

type Seeder struct {
	pg             *persistence.Pg
	ensureDefaults *app.EnsureDefaultsHandler
	listPrices     *app.ListPricesHandler
	createPrice    *app.CreatePriceHandler
	ensurePortal   *app.EnsurePortalConfigurationHandler
	catalog        *Catalog
	identity       *commonconfig.Identity
	logger         *slog.Logger
}

func NewSeeder(
	pg *persistence.Pg,
	ensureDefaults *app.EnsureDefaultsHandler,
	listPrices *app.ListPricesHandler,
	createPrice *app.CreatePriceHandler,
	ensurePortal *app.EnsurePortalConfigurationHandler,
	catalog *Catalog,
	identity *commonconfig.Identity,
	logger *slog.Logger,
) *Seeder {
	return &Seeder{
		pg:             pg,
		ensureDefaults: ensureDefaults,
		listPrices:     listPrices,
		createPrice:    createPrice,
		ensurePortal:   ensurePortal,
		catalog:        catalog,
		identity:       identity,
		logger:         logger,
	}
}

var ProvideSeeder = NewSeeder

// Run is idempotent: a (plan, interval, currency) that already has an active price is skipped.
func (s *Seeder) Run(ctx context.Context) error {
	if err := s.pg.RunMigrations(ctx); err != nil {
		return err
	}
	if err := s.ensureDefaults.Handle(ctx, &app.EnsureDefaults{}); err != nil {
		return err
	}
	prices, err := s.listPrices.Handle(ctx, &app.ListPrices{})
	if err != nil {
		return err
	}
	actor := s.identity.ServiceAccountUserID
	for _, d := range defaultPrices {
		if hasActive(prices, d, s.catalog.Currency) {
			continue
		}
		if err := s.createPrice.Handle(ctx, &app.CreatePrice{
			ID:       uuid.Must(uuid.NewV7()),
			Plan:     d.plan,
			Interval: d.interval,
			Amount:   d.amount,
			Currency: s.catalog.Currency,
			UserID:   actor,
		}); err != nil {
			return err
		}
		s.logger.InfoContext(
			ctx,
			"price created",
			slog.String("plan", d.plan.String()),
			slog.String("interval", d.interval.String()),
		)
	}
	return s.ensurePortal.Handle(ctx, &app.EnsurePortalConfiguration{UserID: actor})
}

func hasActive(prices []app.Price, d defaultPrice, currency string) bool {
	for _, p := range prices {
		if p.Active && p.Plan == d.plan && p.Interval == d.interval && p.Currency == currency {
			return true
		}
	}
	return false
}
