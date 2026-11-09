package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CancelProviderSubscription struct {
	SubscriptionID uuid.UUID
}

type CancelProviderSubscriptionHandler struct {
	uow      domain.UnitOfWork
	subs     domain.SubscriptionRepo
	provider PaymentProvider
}

func NewCancelProviderSubscriptionHandler(
	uow domain.UnitOfWork,
	subs domain.SubscriptionRepo,
	provider PaymentProvider,
) *CancelProviderSubscriptionHandler {
	return &CancelProviderSubscriptionHandler{uow: uow, subs: subs, provider: provider}
}

var ProvideCancelProviderSubscriptionHandler = NewCancelProviderSubscriptionHandler

type CancelProviderSubscriptionCmd commonhandler.Cmd[CancelProviderSubscription]

var _ CancelProviderSubscriptionCmd = (*CancelProviderSubscriptionHandler)(nil)

// Handle cancels a locally canceled subscription at the provider, then records the provider confirmed it.
func (h *CancelProviderSubscriptionHandler) Handle(ctx context.Context, cmd *CancelProviderSubscription) error {
	sub, err := h.subs.GetByID(ctx, cmd.SubscriptionID, false)
	if err != nil {
		return err
	}
	if !sub.AwaitingProviderCancel() {
		return nil
	}
	if err := h.provider.CancelSubscriptionNow(ctx, *sub.ProviderSubscriptionID()); err != nil {
		return err
	}
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		s, err := r.Subscription().GetByID(ctx, cmd.SubscriptionID, true)
		if err != nil {
			return err
		}
		if !s.AwaitingProviderCancel() {
			return nil
		}
		s.MarkProviderSynced(time.Now())
		return r.Subscription().Save(ctx, s)
	})
}
