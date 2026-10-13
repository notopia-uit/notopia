package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type HandlePayerRemoved struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type HandlePayerRemovedHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	settings *Settings
}

func NewHandlePayerRemovedHandler(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, settings *Settings) *HandlePayerRemovedHandler {
	return &HandlePayerRemovedHandler{uow: uow, provider: provider, notifier: notifier, settings: settings}
}

var ProvideHandlePayerRemovedHandler = NewHandlePayerRemovedHandler

type HandlePayerRemovedCmd commonhandler.Cmd[HandlePayerRemoved]

var _ HandlePayerRemovedCmd = (*HandlePayerRemovedHandler)(nil)

func (h *HandlePayerRemovedHandler) Handle(ctx context.Context, cmd *HandlePayerRemoved) error {
	var account *domain.BillingAccount
	var sub *domain.Subscription
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if account, err = r.BillingAccount().GetByWorkspace(ctx, cmd.WorkspaceID, false); err != nil {
			return err
		}
		sub, err = r.Subscription().GetCurrentByWorkspace(ctx, cmd.WorkspaceID, false)
		return err
	}); err != nil {
		return err
	}
	if account == nil || account.Payer().ID != cmd.UserID || account.PayerMissingSince() != nil {
		return nil
	}
	recurring := sub != nil && sub.ProviderSubscriptionID() != nil
	if recurring {
		if err := h.provider.SetCancelAtPeriodEnd(ctx, *sub.ProviderSubscriptionID(), true); err != nil {
			return err
		}
	}
	if err := h.provider.DetachPaymentMethods(ctx, account.ProviderCustomerID(), ""); err != nil {
		return err
	}
	now, actor := time.Now(), h.settings.ServiceAccountUserID
	err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		a, err := r.BillingAccount().GetByWorkspace(ctx, cmd.WorkspaceID, true)
		if err != nil {
			return err
		}
		a.MarkPayerMissing(now, actor)
		if err := r.BillingAccount().Save(ctx, a); err != nil {
			return err
		}
		if !recurring {
			return nil
		}
		s, err := r.Subscription().GetByID(ctx, sub.ID(), true)
		if err != nil {
			return err
		}
		if err := s.ScheduleCancel(now, actor); err != nil {
			return err
		}
		return r.Subscription().Save(ctx, s)
	})
	if err == nil && sub != nil {
		notifySubscription(ctx, h.notifier, sub)
	}
	return err
}
