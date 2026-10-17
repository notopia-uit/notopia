package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type RefundRunner struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func NewRefundRunner(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *RefundRunner {
	return &RefundRunner{uow: uow, provider: provider, settings: settings}
}

var ProvideRefundRunner = NewRefundRunner

// Run submits every pending refund of the subscription that has no provider id yet.
func (r *RefundRunner) Run(ctx context.Context, subscriptionID uuid.UUID) error {
	var pending []*domain.Refund
	if err := r.uow.Execute(ctx, func(reg domain.RepoRegistry) error {
		sub, err := reg.Subscription().GetByID(ctx, subscriptionID, false)
		if err != nil {
			return err
		}
		for _, rf := range sub.Refunds() {
			if rf.Status() == domain.RefundStatusPending && rf.ProviderRefundID() == nil {
				pending = append(pending, rf)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, rf := range pending {
		res, err := r.provider.CreateRefund(ctx, rf.ID(), rf.PaymentRef(), rf.Amount())
		if hasCode(err, errs.CodeInvalid) {
			slog.ErrorContext(
				ctx, "refund rejected by the provider; marking it failed",
				slog.String("refund_id", rf.ID().String()),
				slog.Any("error", err),
			)
			if err := r.failRefund(ctx, subscriptionID, rf.ID(), err.Error()); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			slog.ErrorContext(
				ctx, "refund submission failed; cron will retry",
				slog.String("refund_id", rf.ID().String()),
				slog.Any("error", err),
			)
			continue
		}
		if err := r.uow.Execute(ctx, func(reg domain.RepoRegistry) error {
			sub, err := reg.Subscription().GetByID(ctx, subscriptionID, true)
			if err != nil {
				return err
			}
			if err := sub.UpdateRefund(rf.ID(), res.ProviderRefundID, res.Status, res.FailureMessage, time.Now(), r.settings.ServiceAccountUserID); err != nil {
				return err
			}
			return reg.Subscription().Save(ctx, sub)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *RefundRunner) failRefund(ctx context.Context, subscriptionID, refundID uuid.UUID, message string) error {
	return r.uow.Execute(ctx, func(reg domain.RepoRegistry) error {
		sub, err := reg.Subscription().GetByID(ctx, subscriptionID, true)
		if err != nil {
			return err
		}
		if err := sub.FailRefund(refundID, message, time.Now(), r.settings.ServiceAccountUserID); err != nil {
			return err
		}
		return reg.Subscription().Save(ctx, sub)
	})
}
