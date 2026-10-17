package billingfixtures

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

// Users come from deploy/compose/infrastructure/authentik-blueprints/users.yaml.
// Provider ids are fake: Stripe actions on this workspace fail, entitlements work.
var (
	workspaceID = uuid.MustParse("00000000-0000-4000-8000-000000000110")
	payer       = domain.Payer{
		ID:    "110",
		Email: "user@notopia.com",
		Name:  "User",
	}
)

const (
	actor                  = "110"
	currency               = "vnd"
	fakeProviderCustomerID = "cus_fixture_110"
	fakeProviderPaymentID  = "pi_fixture_110"
)

type Fixtures struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewFixtures(uow domain.UnitOfWork, logger *slog.Logger) *Fixtures {
	return &Fixtures{
		uow:    uow,
		logger: logger,
	}
}

var ProvideFixtures = NewFixtures

// Run is idempotent: skips when the workspace already has a subscription.
func (f *Fixtures) Run(ctx context.Context) error {
	now := time.Now()
	return f.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		existing, err := r.Subscription().GetCurrentByWorkspace(ctx, workspaceID, true)
		if err != nil {
			return err
		}
		if existing != nil {
			f.logger.InfoContext(ctx, "workspace already has a subscription, skipped", slog.String("workspaceId", workspaceID.String()))
			return nil
		}
		price, err := r.Price().GetActive(ctx, domain.PlanTierPro, domain.IntervalLifetime, currency, false)
		if err != nil {
			return err
		}
		if price == nil {
			return errors.New("pro lifetime price missing: run just seed-billing first")
		}
		account := domain.OpenBillingAccount(workspaceID, fakeProviderCustomerID, payer, now, actor)
		if err := r.BillingAccount().Save(ctx, account); err != nil {
			return err
		}
		sub, err := domain.NewLifetimeSubscription(uuid.Must(uuid.NewV7()), workspaceID, price, fakeProviderPaymentID, now, now, actor)
		if err != nil {
			return err
		}
		if err := r.Subscription().Save(ctx, sub); err != nil {
			return err
		}
		f.logger.InfoContext(ctx, "lifetime subscription created", slog.String("workspaceId", workspaceID.String()))
		return nil
	})
}
