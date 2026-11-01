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
	subs    domain.SubscriptionRepo
	sync    *SyncSubscriptionHandler
	refunds *RefundRunner
}

func NewReconcileSubscriptionsHandler(
	subs domain.SubscriptionRepo,
	sync *SyncSubscriptionHandler,
	refunds *RefundRunner,
) *ReconcileSubscriptionsHandler {
	return &ReconcileSubscriptionsHandler{subs: subs, sync: sync, refunds: refunds}
}

var ProvideReconcileSubscriptionsHandler = NewReconcileSubscriptionsHandler

type ReconcileSubscriptionsCmd commonhandler.Cmd[ReconcileSubscriptions]

var _ ReconcileSubscriptionsCmd = (*ReconcileSubscriptionsHandler)(nil)

// Handle re-syncs subscriptions whose provider state may be stale and retries refunds
// that were never submitted. Locally canceled subscriptions are synced too, which retries a failed provider cancel. Per-item failures are logged; the next tick retries them.
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
	for _, s := range append(stale, canceled...) {
		if s.ProviderSubscriptionID() == nil {
			continue
		}
		cmd := &SyncSubscription{EventID: "", EventType: "", ProviderSubscriptionID: *s.ProviderSubscriptionID()}
		if err := h.sync.Handle(ctx, cmd); err != nil {
			slog.WarnContext(
				ctx, "reconcile failed",
				slog.String("subscription_id", s.ID().String()),
				slog.Any("error", err),
			)
		}
	}
	for _, id := range refundSubs {
		if err := h.refunds.Run(ctx, id); err != nil {
			slog.WarnContext(
				ctx, "refund retry failed",
				slog.String("subscription_id", id.String()),
				slog.Any("error", err),
			)
		}
	}
	return nil
}
