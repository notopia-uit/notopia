package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

const reconcileBatchSize = 100

type ReconcileSubscriptions struct{}

type ReconcileSubscriptionsHandler struct {
	subs           domain.SubscriptionRepo
	sync           *SyncSubscriptionHandler
	cancelProvider *CancelProviderSubscriptionHandler
	submitRefund   *SubmitRefundHandler
}

func NewReconcileSubscriptionsHandler(
	subs domain.SubscriptionRepo,
	sync *SyncSubscriptionHandler,
	cancelProvider *CancelProviderSubscriptionHandler,
	submitRefund *SubmitRefundHandler,
) *ReconcileSubscriptionsHandler {
	return &ReconcileSubscriptionsHandler{subs: subs, sync: sync, cancelProvider: cancelProvider, submitRefund: submitRefund}
}

var ProvideReconcileSubscriptionsHandler = NewReconcileSubscriptionsHandler

type ReconcileSubscriptionsCmd commonhandler.Cmd[ReconcileSubscriptions]

var _ ReconcileSubscriptionsCmd = (*ReconcileSubscriptionsHandler)(nil)

// Handle is the safety net behind the outbox consumers: it re-syncs subscriptions whose provider state may be
// stale, retries provider cancels of locally canceled subscriptions and submits refunds still pending.
// Per-item failures are logged; the next tick retries them.
func (h *ReconcileSubscriptionsHandler) Handle(ctx context.Context, _ *ReconcileSubscriptions) error {
	stale, err := h.subs.ListNeedingReconcile(ctx, time.Now(), reconcileBatchSize)
	if err != nil {
		return err
	}
	canceled, err := h.subs.ListCanceledAwaitingProviderCancel(ctx, reconcileBatchSize)
	if err != nil {
		return err
	}
	refundSubs, err := h.subs.ListIDsWithUnsubmittedRefunds(ctx, reconcileBatchSize)
	if err != nil {
		return err
	}
	for _, s := range stale {
		if s.ProviderSubscriptionID() == nil {
			continue
		}
		cmd := &SyncSubscription{EventID: "", EventType: "", ProviderSubscriptionID: *s.ProviderSubscriptionID()}
		if err := h.sync.Handle(ctx, cmd); err != nil {
			slog.WarnContext(
				ctx, "reconcile sync failed",
				slog.String("subscription_id", s.ID().String()),
				slog.Any("error", err),
			)
		}
	}
	for _, s := range canceled {
		if err := h.cancelProvider.Handle(ctx, &CancelProviderSubscription{SubscriptionID: s.ID()}); err != nil {
			slog.WarnContext(
				ctx, "reconcile provider cancel failed",
				slog.String("subscription_id", s.ID().String()),
				slog.Any("error", err),
			)
		}
	}
	for _, id := range refundSubs {
		sub, err := h.subs.GetByID(ctx, id, false)
		if err != nil {
			slog.WarnContext(
				ctx, "reconcile refund lookup failed",
				slog.String("subscription_id", id.String()),
				slog.Any("error", err),
			)
			continue
		}
		for _, rf := range sub.Refunds() {
			if !rf.AwaitingSubmission() {
				continue
			}
			if err := h.submitRefund.Handle(ctx, &SubmitRefund{SubscriptionID: id, RefundID: rf.ID()}); err != nil {
				slog.WarnContext(
					ctx, "reconcile refund submission failed",
					slog.String("subscription_id", id.String()),
					slog.String("refund_id", rf.ID().String()),
					slog.Any("error", err),
				)
			}
		}
	}
	return nil
}
