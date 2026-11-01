package app

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type SyncSubscription struct {
	EventID                string // empty when called by the reconcile cron
	EventType              string
	ProviderSubscriptionID string
}

type SyncSubscriptionHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func NewSyncSubscriptionHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *SyncSubscriptionHandler {
	return &SyncSubscriptionHandler{uow: uow, provider: provider, settings: settings}
}

var ProvideSyncSubscriptionHandler = NewSyncSubscriptionHandler

type SyncSubscriptionCmd commonhandler.Cmd[SyncSubscription]

var _ SyncSubscriptionCmd = (*SyncSubscriptionHandler)(nil)

func (h *SyncSubscriptionHandler) Handle(ctx context.Context, cmd *SyncSubscription) error {
	snap, err := h.provider.GetSubscriptionSnapshot(ctx, cmd.ProviderSubscriptionID)
	if err != nil {
		return err
	}
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if cmd.EventID != "" {
			if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
				return err
			}
		}
		sub, err := r.Subscription().GetByProviderSubscriptionID(ctx, cmd.ProviderSubscriptionID, true)
		if err != nil || sub == nil {
			return err // unknown yet: checkout.session.completed creates it
		}
		if !sub.IsCurrent() {
			// canceled locally (cancel-now / duplicate / replaced): never resurrect; while the provider still
			// runs it, CancelProviderSubscription (outbox event, reconcile) cancels it there
			if snap.Status() == domain.SubscriptionStatusCanceled || snap.Status() == domain.SubscriptionStatusIncompleteExpired {
				sub.MarkProviderSynced(time.Now())
				return r.Subscription().Save(ctx, sub)
			}
			return nil
		}
		price, err := r.Price().GetByProviderPriceID(ctx, snap.ProviderPriceID())
		if err != nil {
			return err
		}
		sub.ApplySnapshot(snap, price, time.Now(), h.settings.ServiceAccountUserID)
		return r.Subscription().Save(ctx, sub)
	})
}
