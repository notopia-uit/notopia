package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

// CancelNowCause is who asked for the cancel-now; it decides the refund reason and whether a refund is required.
type CancelNowCause uint8

const (
	CancelNowCauseUnspecified CancelNowCause = iota
	CancelNowCauseUser
	CancelNowCauseWorkspaceDeleted
)

type SubscriptionCanceller struct {
	uow      domain.UnitOfWork
	subs     domain.SubscriptionRepo
	policies domain.BillingPolicyRepo
	provider PaymentProvider
}

func NewSubscriptionCanceller(
	uow domain.UnitOfWork,
	subs domain.SubscriptionRepo,
	policies domain.BillingPolicyRepo,
	provider PaymentProvider,
) *SubscriptionCanceller {
	return &SubscriptionCanceller{uow: uow, subs: subs, policies: policies, provider: provider}
}

var ProvideSubscriptionCanceller = NewSubscriptionCanceller

func (c *SubscriptionCanceller) load(ctx context.Context, ws uuid.UUID) (*domain.Subscription, *domain.BillingPolicy, error) {
	sub, err := c.subs.GetCurrentByWorkspace(ctx, ws, false)
	if err != nil {
		return nil, nil, err
	}
	if sub == nil {
		return nil, nil, errs.New(errs.CodeNoActiveSubscription, "no active subscription", nil)
	}
	policy, err := c.policies.Get(ctx, false)
	if err != nil {
		return nil, nil, err
	}
	return sub, policy, nil
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

// CancelNow cancels locally and requests the refunds in one transaction; the outbox events drive
// the provider cancel (CancelProviderSubscription) and the refund submission (SubmitRefund).
// A deleted workspace cancels even when no refund is due (lifetime past its window).
// It is idempotent: a subscription already canceled by a concurrent call is left as is.
func (c *SubscriptionCanceller) CancelNow(ctx context.Context, ws uuid.UUID, cause CancelNowCause, actor string) error {
	sub, policy, err := c.load(ctx, ws)
	if err != nil {
		return err
	}
	var reason domain.RefundReason
	switch cause {
	case CancelNowCauseUser:
		reason = domain.RefundReasonCancelNow
		if sub.Interval() == domain.IntervalLifetime {
			reason = domain.RefundReasonLifetimeWindow
		}
	case CancelNowCauseWorkspaceDeleted:
		reason = domain.RefundReasonWorkspaceDeleted
	case CancelNowCauseUnspecified:
		return errs.NewInternal(fmt.Sprintf("unknown cancel-now cause %d", cause), nil)
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
		if cause != CancelNowCauseWorkspaceDeleted || !isRefundWindowExpired {
			return err
		}
		if quote, err = domain.NewRefundQuote(nil, 0, 0); err != nil {
			return err
		}
	}
	return c.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		s, err := r.Subscription().GetByID(ctx, sub.ID(), true)
		if err != nil {
			return err
		}
		if !s.IsCurrent() {
			return nil
		}
		if _, err := s.CancelNow(now, reason, quote, actor); err != nil {
			return err
		}
		return r.Subscription().Save(ctx, s)
	})
}
