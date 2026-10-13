package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Canceller struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	refunds  *RefundRunner
}

func NewCanceller(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, refunds *RefundRunner) *Canceller {
	return &Canceller{uow: uow, provider: provider, notifier: notifier, refunds: refunds}
}

var ProvideCanceller = NewCanceller

func (c *Canceller) load(ctx context.Context, ws uuid.UUID) (*domain.Subscription, *domain.BillingPolicy, error) {
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

func (c *Canceller) payments(ctx context.Context, sub *domain.Subscription) ([]domain.PaymentInfo, error) {
	if sub.Interval() == domain.IntervalLifetime {
		p, err := c.provider.PaymentForIntent(ctx, *sub.ProviderPaymentID())
		return []domain.PaymentInfo{p}, err
	}
	if sub.CurrentPeriodStart() == nil || sub.CurrentPeriodEnd() == nil {
		return nil, errs.NewInternal("subscription has no period", nil)
	}
	return c.provider.PeriodPayments(ctx, *sub.ProviderSubscriptionID(), *sub.CurrentPeriodStart(), *sub.CurrentPeriodEnd())
}

func (c *Canceller) Quote(ctx context.Context, ws uuid.UUID) (domain.RefundQuote, error) {
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

// CancelNow cancels at the provider first, then locally, then submits the refund.
// refundOptional=true (workspace deleted) cancels even when no refund is due (lifetime past its window).
func (c *Canceller) CancelNow(ctx context.Context, ws uuid.UUID, reason domain.RefundReason, actor string, refundOptional bool) error {
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
		if !refundOptional || !hasCode(err, errs.CodeRefundWindowExpired) {
			return err
		}
		var noRefund domain.RefundQuote
		quote = noRefund
	}
	if sub.ProviderSubscriptionID() != nil {
		if err := c.provider.CancelSubscriptionNow(ctx, *sub.ProviderSubscriptionID()); err != nil {
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
	notifySubscription(ctx, c.notifier, saved)
	return c.refunds.Run(ctx, saved.ID())
}
