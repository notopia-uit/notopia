package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type SubscriptionCanceller struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	refunds  *RefundRunner
}

func NewSubscriptionCanceller(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, refunds *RefundRunner) *SubscriptionCanceller {
	return &SubscriptionCanceller{uow: uow, provider: provider, notifier: notifier, refunds: refunds}
}

var ProvideSubscriptionCanceller = NewSubscriptionCanceller

func (c *SubscriptionCanceller) load(ctx context.Context, ws uuid.UUID) (*domain.Subscription, *domain.BillingPolicy, error) {
	var sub *domain.Subscription
	var policy *domain.BillingPolicy
	err := c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if sub, err = r.Subscription().GetCurrentByWorkspace(ctx, ws, false); err != nil {
			return err
		}
		policy, err = r.BillingPolicy().Get(ctx, false)
		return err
	})
	if err == nil && sub == nil {
		err = errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	return sub, policy, err
}

func (c *SubscriptionCanceller) payments(ctx context.Context, sub *domain.Subscription) ([]domain.PaymentInfo, error) {
	if sub.Interval() == domain.IntervalLifetime {
		p, err := c.provider.PaymentForIntent(ctx, *sub.ProviderPaymentID())
		return []domain.PaymentInfo{p}, err
	}
	if sub.CurrentPeriodStart() == nil || sub.CurrentPeriodEnd() == nil {
		return nil, errs.NewInternal("subscription has no period", nil)
	}
	return c.provider.PeriodPayments(ctx, *sub.ProviderSubscriptionID(), *sub.CurrentPeriodStart(), *sub.CurrentPeriodEnd())
}

func (c *SubscriptionCanceller) Quote(ctx context.Context, ws uuid.UUID) (domain.RefundQuote, error) {
	sub, policy, err := c.load(ctx, ws)
	if err != nil {
		return domain.RefundQuote{}, err
	}
	payments, err := c.payments(ctx, sub)
	if err != nil {
		return domain.RefundQuote{}, err
	}
	return sub.QuoteCancelNow(time.Now(), policy, payments)
}

// CancelNow cancels locally and commits, then cancels at the provider (retried by reconcile on
// failure), then submits the refunds.
// refundOptional=true (workspace deleted) cancels even when no refund is due (lifetime past its window).
func (c *SubscriptionCanceller) CancelNow(ctx context.Context, ws uuid.UUID, reason domain.RefundReason, actor string, refundOptional bool) error {
	sub, policy, err := c.load(ctx, ws)
	if err != nil {
		return err
	}
	payments, err := c.payments(ctx, sub)
	if err != nil {
		return err
	}
	now := time.Now()
	quote, err := sub.QuoteCancelNow(now, policy, payments)
	if err != nil {
		var errsErr errs.Error
		isRefundWindowExpired := errors.As(err, &errsErr) && errsErr.Code() == errs.CodeRefundWindowExpired
		if !refundOptional || !isRefundWindowExpired {
			return err
		}
		if quote, err = domain.NewRefundQuote(nil, 0, 0); err != nil {
			return err
		}
	}
	var saved *domain.Subscription
	if err := c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		s, err := r.Subscription().GetByID(ctx, sub.ID(), true)
		if err != nil || !s.IsCurrent() {
			return err
		}
		if _, err := s.CancelNow(now, reason, quote, actor); err != nil {
			return err
		}
		saved = s
		return r.Subscription().Save(ctx, s)
	}); err != nil || saved == nil {
		return err
	}
	if saved.ProviderSubscriptionID() != nil {
		providerSubscriptionID := *saved.ProviderSubscriptionID()
		if err := c.provider.CancelSubscriptionNow(ctx, providerSubscriptionID); err != nil {
			slog.ErrorContext(
				ctx, "cancelling subscription at the provider failed; reconcile will retry",
				slog.String("provider_subscription_id", providerSubscriptionID),
				slog.Any("error", err),
			)
		}
	}
	notifySubscription(ctx, c.notifier, saved)
	return c.refunds.Run(ctx, saved.ID())
}
