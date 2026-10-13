package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CompleteCheckout struct {
	EventID   string
	SessionID string
}

type CompleteCheckoutHandler struct {
	uow      domain.UnitOfWork
	provider PaymentProvider
	notifier Notifier
	settings *Settings
	refunds  *RefundRunner
}

func NewCompleteCheckoutHandler(uow domain.UnitOfWork, provider PaymentProvider, notifier Notifier, settings *Settings, refunds *RefundRunner) *CompleteCheckoutHandler {
	return &CompleteCheckoutHandler{uow: uow, provider: provider, notifier: notifier, settings: settings, refunds: refunds}
}

var ProvideCompleteCheckoutHandler = NewCompleteCheckoutHandler

type CompleteCheckoutCmd commonhandler.Cmd[CompleteCheckout]

var _ CompleteCheckoutCmd = (*CompleteCheckoutHandler)(nil)

const eventCheckoutCompleted = "checkout.session.completed"

func (h *CompleteCheckoutHandler) Handle(ctx context.Context, cmd *CompleteCheckout) error {
	session, err := h.provider.GetCheckoutSession(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if !session.Complete {
		return h.markProcessed(ctx, cmd)
	}
	switch session.Purpose {
	case CheckoutPurposeSubscribe:
		return h.subscribe(ctx, cmd, session)
	case CheckoutPurposeLifetime:
		return h.lifetime(ctx, cmd, session)
	case CheckoutPurposePayerChange:
		return h.payerChange(ctx, cmd, session)
	case CheckoutPurposeUnspecified:
		slog.WarnContext(
			ctx, "ignoring checkout session without a known purpose",
			slog.String("session_id", session.ID),
		)
	}
	return h.markProcessed(ctx, cmd)
}

func (h *CompleteCheckoutHandler) markProcessed(ctx context.Context, cmd *CompleteCheckout) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		_, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted)
		return err
	})
}

func (h *CompleteCheckoutHandler) currentSub(ctx context.Context, ws uuid.UUID) (*domain.Subscription, error) {
	var cur *domain.Subscription
	err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		cur, err = r.Subscription().GetCurrentByWorkspace(ctx, ws, false)
		return err
	})
	return cur, err
}

func (h *CompleteCheckoutHandler) setPayer(ctx context.Context, r domain.RepoRegistry, ws uuid.UUID, payer domain.Payer, now time.Time) error {
	account, err := r.BillingAccount().GetByWorkspace(ctx, ws, true)
	if err != nil || account == nil {
		return err
	}
	if account.Payer().ID != payer.ID || account.PayerMissingSince() != nil {
		account.ChangePayer(payer, now, payer.ID)
		return r.BillingAccount().Save(ctx, account)
	}
	return nil
}

func (h *CompleteCheckoutHandler) subscribe(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	snap, err := h.provider.GetSubscriptionSnapshot(ctx, s.ProviderSubscriptionID)
	if err != nil {
		return err
	}
	cur, err := h.currentSub(ctx, s.WorkspaceID)
	if err != nil {
		return err
	}
	duplicate := cur != nil && (cur.ProviderSubscriptionID() == nil || *cur.ProviderSubscriptionID() != s.ProviderSubscriptionID)
	var payments []domain.PaymentInfo
	if duplicate {
		if payments, err = h.provider.PeriodPayments(ctx, s.ProviderSubscriptionID, snap.CurrentPeriodStart, snap.CurrentPeriodEnd); err != nil {
			return err
		}
		if err := h.provider.CancelSubscriptionNow(ctx, s.ProviderSubscriptionID); err != nil {
			return err
		}
	}

	now, actor := time.Now(), h.settings.ServiceAccountUserID
	var saved *domain.Subscription
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted); err != nil || !first {
			return err
		}
		if existing, err := r.Subscription().GetByProviderSubscriptionID(ctx, s.ProviderSubscriptionID, true); err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetByProviderPriceID(ctx, snap.ProviderPriceID)
		if err != nil {
			return err
		}
		sub, err := domain.NewSubscription(uuid.Must(uuid.NewV7()), s.WorkspaceID, price, snap, now, actor)
		if err != nil {
			return err
		}
		if duplicate {
			if _, err := sub.CancelNow(now, domain.RefundReasonDuplicateCheckout, domain.FullRefundQuote(payments...), actor); err != nil {
				return err
			}
		} else if err := h.setPayer(ctx, r, s.WorkspaceID, s.Payer, now); err != nil {
			return err
		}
		saved = sub
		return r.Subscription().Save(ctx, sub)
	})
	if err != nil || saved == nil {
		return err
	}
	if duplicate {
		return h.refunds.Run(ctx, saved.ID())
	}
	notifySubscription(ctx, h.notifier, saved)
	return nil
}

func (h *CompleteCheckoutHandler) lifetime(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	cur, err := h.currentSub(ctx, s.WorkspaceID)
	if err != nil {
		return err
	}
	var replaced []domain.PaymentInfo
	replacing := cur != nil && cur.Interval() != domain.IntervalLifetime
	duplicate := cur != nil && cur.Interval() == domain.IntervalLifetime &&
		(cur.ProviderPaymentID() == nil || *cur.ProviderPaymentID() != s.Payment.PaymentRef)
	if replacing {
		if replaced, err = h.provider.PeriodPayments(ctx, *cur.ProviderSubscriptionID(), *cur.CurrentPeriodStart(), *cur.CurrentPeriodEnd()); err != nil {
			return err
		}
		if err := h.provider.CancelSubscriptionNow(ctx, *cur.ProviderSubscriptionID()); err != nil {
			return err
		}
	}

	now, actor := time.Now(), h.settings.ServiceAccountUserID
	var toRefund []uuid.UUID
	var saved *domain.Subscription
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted); err != nil || !first {
			return err
		}
		if existing, err := r.Subscription().GetByProviderPaymentID(ctx, s.Payment.PaymentRef, true); err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetByID(ctx, s.PriceID, false)
		if err != nil {
			return err
		}
		policy, err := r.BillingPolicy().Get(ctx, false)
		if err != nil {
			return err
		}
		if replacing {
			old, err := r.Subscription().GetByID(ctx, cur.ID(), true)
			if err != nil {
				return err
			}
			quote, err := old.QuoteCancelNow(now, policy, replaced)
			if err != nil {
				return err
			}
			if _, err := old.CancelNow(now, domain.RefundReasonReplacedByLifetime, quote, actor); err != nil {
				return err
			}
			if err := r.Subscription().Save(ctx, old); err != nil {
				return err
			}
			toRefund = append(toRefund, old.ID())
		}
		sub, err := domain.NewLifetimeSubscription(uuid.Must(uuid.NewV7()), s.WorkspaceID, price, s.Payment.PaymentRef, s.Payment.PaidAt, now, actor)
		if err != nil {
			return err
		}
		if duplicate {
			if _, err := sub.CancelNow(now, domain.RefundReasonDuplicateCheckout, domain.FullRefundQuote(s.Payment), actor); err != nil {
				return err
			}
			toRefund = append(toRefund, sub.ID())
		} else if err := h.setPayer(ctx, r, s.WorkspaceID, s.Payer, now); err != nil {
			return err
		}
		saved = sub
		return r.Subscription().Save(ctx, sub)
	})
	if err != nil || saved == nil {
		return err
	}
	for _, id := range toRefund {
		if err := h.refunds.Run(ctx, id); err != nil {
			return err
		}
	}
	if !duplicate {
		notifySubscription(ctx, h.notifier, saved)
	}
	return nil
}

func (h *CompleteCheckoutHandler) payerChange(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	var account *domain.BillingAccount
	cur, err := h.currentSub(ctx, s.WorkspaceID)
	if err != nil {
		return err
	}
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		account, err = r.BillingAccount().GetByWorkspace(ctx, s.WorkspaceID, false)
		return err
	}); err != nil || account == nil {
		return err
	}
	resume := account.PayerMissingSince() != nil && cur != nil && cur.CancelAtPeriodEnd() && cur.ProviderSubscriptionID() != nil
	var providerSubID *string
	if cur != nil {
		providerSubID = cur.ProviderSubscriptionID()
	}
	if err := h.provider.SetDefaultPaymentMethod(ctx, account.ProviderCustomerID(), s.PaymentMethodID, providerSubID); err != nil {
		return err
	}
	if err := h.provider.DetachPaymentMethods(ctx, account.ProviderCustomerID(), s.PaymentMethodID); err != nil {
		return err
	}
	if err := h.provider.UpdateCustomerContact(ctx, account.ProviderCustomerID(), s.Payer); err != nil {
		return err
	}
	if resume {
		if err := h.provider.SetCancelAtPeriodEnd(ctx, *cur.ProviderSubscriptionID(), false); err != nil {
			return err
		}
	}

	now := time.Now()
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, eventCheckoutCompleted); err != nil || !first {
			return err
		}
		if err := h.setPayer(ctx, r, s.WorkspaceID, s.Payer, now); err != nil {
			return err
		}
		if !resume {
			return nil
		}
		sub, err := r.Subscription().GetByID(ctx, cur.ID(), true)
		if err != nil {
			return err
		}
		sub.ResumeCancel(now, s.Payer.ID)
		return r.Subscription().Save(ctx, sub)
	})
	if err == nil && cur != nil {
		notifySubscription(ctx, h.notifier, cur)
	}
	return err
}
