package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

const reconcileBatchSize = 100

type ReconcileSubscriptions struct{}

type ReconcileSubscriptionsHandler struct {
	uow     domain.UnitOfWork
	sync    *SyncSubscriptionHandler
	refunds *RefundRunner
}

func NewReconcileSubscriptionsHandler(
	uow domain.UnitOfWork,
	sync *SyncSubscriptionHandler,
	refunds *RefundRunner,
) *ReconcileSubscriptionsHandler {
	return &ReconcileSubscriptionsHandler{uow: uow, sync: sync, refunds: refunds}
}

var ProvideReconcileSubscriptionsHandler = NewReconcileSubscriptionsHandler

type ReconcileSubscriptionsCmd commonhandler.Cmd[ReconcileSubscriptions]

var _ ReconcileSubscriptionsCmd = (*ReconcileSubscriptionsHandler)(nil)

// Handle re-syncs subscriptions whose provider state may be stale and retries refunds
// that were never submitted. Per-item failures are logged; the next tick retries them.
func (h *ReconcileSubscriptionsHandler) Handle(ctx context.Context, _ *ReconcileSubscriptions) error {
	var stale []*domain.Subscription
	var refundSubs []uuid.UUID
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if stale, err = r.Subscription().ListNeedingReconcile(ctx, time.Now(), reconcileBatchSize); err != nil {
			return err
		}
		refundSubs, err = r.Subscription().ListIDsWithUnsubmittedRefunds(ctx, reconcileBatchSize)
		return err
	}); err != nil {
		return err
	}
	for _, s := range stale {
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
