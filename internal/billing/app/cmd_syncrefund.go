package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type SyncRefund struct {
	EventID          string
	EventType        string
	ProviderRefundID string
}

type SyncRefundHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	settings *Settings
}

func NewSyncRefundHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *SyncRefundHandler {
	return &SyncRefundHandler{uow: uow, provider: provider, settings: settings}
}

var ProvideSyncRefundHandler = NewSyncRefundHandler

type SyncRefundCmd commonhandler.Cmd[SyncRefund]

var _ SyncRefundCmd = (*SyncRefundHandler)(nil)

func (h *SyncRefundHandler) Handle(ctx context.Context, cmd *SyncRefund) error {
	res, err := h.provider.GetRefund(ctx, cmd.ProviderRefundID)
	if err != nil {
		return err
	}
	now, actor := time.Now(), h.settings.ServiceAccountUserID
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
			return err
		}
		if res.RefundID != nil {
			sub, err := r.Subscription().GetByRefundID(ctx, *res.RefundID, true)
			if err != nil || sub == nil {
				return err
			}
			if err := sub.UpdateRefund(*res.RefundID, res.ProviderRefundID, res.Status, res.FailureMessage, now, actor); err != nil {
				return err
			}
			return r.Subscription().Save(ctx, sub)
		}
		sub, err := r.Subscription().GetByProviderPaymentID(ctx, res.PaymentRef, true)
		if err != nil {
			return err
		}
		if sub == nil && res.ProviderSubscriptionID != "" {
			if sub, err = r.Subscription().GetByProviderSubscriptionID(ctx, res.ProviderSubscriptionID, true); err != nil {
				return err
			}
		}
		if sub == nil {
			slog.InfoContext(ctx, "ignoring external refund of an unknown payment", slog.String("refund", res.ProviderRefundID))
			return nil
		}
		sub.RecordExternalRefund(res.ProviderRefundID, res.PaymentRef, res.Amount, res.Status, now, actor)
		return r.Subscription().Save(ctx, sub)
	})
}
