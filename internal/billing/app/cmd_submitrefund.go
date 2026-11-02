package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type SubmitRefund struct {
	SubscriptionID uuid.UUID
	RefundID       uuid.UUID
}

type SubmitRefundHandler struct {
	uow      domain.UnitOfWork
	subs     domain.SubscriptionRepo
	provider PaymentProvider
	settings *Settings
}

func NewSubmitRefundHandler(
	uow domain.UnitOfWork,
	subs domain.SubscriptionRepo,
	provider PaymentProvider,
	settings *Settings,
) *SubmitRefundHandler {
	return &SubmitRefundHandler{uow: uow, subs: subs, provider: provider, settings: settings}
}

var ProvideSubmitRefundHandler = NewSubmitRefundHandler

type SubmitRefundCmd commonhandler.Cmd[SubmitRefund]

var _ SubmitRefundCmd = (*SubmitRefundHandler)(nil)

// Handle submits a pending refund to the provider, keyed by the refund id so retries never refund twice.
// A refund the provider rejects as invalid is marked failed; any other provider error is returned for a retry.
func (h *SubmitRefundHandler) Handle(ctx context.Context, cmd *SubmitRefund) error {
	sub, err := h.subs.GetByID(ctx, cmd.SubscriptionID, false)
	if err != nil {
		return err
	}
	rf := sub.Refund(cmd.RefundID)
	if rf == nil {
		return errs.NewNotFound("refund " + cmd.RefundID.String() + " not found")
	}
	if !rf.AwaitingSubmission() {
		return nil
	}
	res, err := h.provider.CreateRefund(ctx, rf.ID(), rf.PaymentRef(), rf.Amount())
	var errsErr errs.Error
	if errors.As(err, &errsErr) && errsErr.Code() == errs.CodeInvalid {
		slog.ErrorContext(
			ctx, "refund rejected by the provider; marking it failed",
			slog.String("refund_id", rf.ID().String()),
			slog.Any("error", err),
		)
		return h.update(ctx, cmd, func(s *domain.Subscription, now time.Time) error {
			return s.FailRefund(cmd.RefundID, err.Error(), now, h.settings.ServiceAccountUserID)
		})
	}
	if err != nil {
		return err
	}
	return h.update(ctx, cmd, func(s *domain.Subscription, now time.Time) error {
		return s.UpdateRefund(cmd.RefundID, res.ProviderRefundID, res.Status, res.FailureMessage, now, h.settings.ServiceAccountUserID)
	})
}

func (h *SubmitRefundHandler) update(ctx context.Context, cmd *SubmitRefund, apply func(*domain.Subscription, time.Time) error) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		s, err := r.Subscription().GetByID(ctx, cmd.SubscriptionID, true)
		if err != nil {
			return err
		}
		// the refund webhook (SyncRefund) may have recorded the provider refund meanwhile; it read live state
		if rf := s.Refund(cmd.RefundID); rf == nil || !rf.AwaitingSubmission() {
			return nil
		}
		if err := apply(s, time.Now()); err != nil {
			return err
		}
		return r.Subscription().Save(ctx, s)
	})
}
