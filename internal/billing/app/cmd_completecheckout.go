package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CompleteCheckout struct {
	EventID   string
	EventType string
	SessionID string
}

type CompleteCheckoutHandler struct {
	uow      domain.UnitOfWork
	subs     domain.SubscriptionRepo
	accounts domain.BillingAccountRepo
	provider PaymentProvider
	settings *Settings
}

func NewCompleteCheckoutHandler(
	uow domain.UnitOfWork,
	subs domain.SubscriptionRepo,
	accounts domain.BillingAccountRepo,
	provider PaymentProvider,
	settings *Settings,
) *CompleteCheckoutHandler {
	return &CompleteCheckoutHandler{uow: uow, subs: subs, accounts: accounts, provider: provider, settings: settings}
}

var ProvideCompleteCheckoutHandler = NewCompleteCheckoutHandler

type CompleteCheckoutCmd commonhandler.Cmd[CompleteCheckout]

var _ CompleteCheckoutCmd = (*CompleteCheckoutHandler)(nil)

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
		_, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType)
		return err
	})
}

// lockAccount locks the workspace's billing account, opened by StartCheckout before any checkout. The lock
// serializes the checkouts of one workspace, so each one sees the subscription a concurrent one committed.
func (h *CompleteCheckoutHandler) lockAccount(ctx context.Context, r domain.RepoRegistry, ws uuid.UUID) (*domain.BillingAccount, error) {
	account, err := r.BillingAccount().GetByWorkspace(ctx, ws, true)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errs.New(errs.CodeNoBillingAccount, "checkout completed for a workspace without a billing account", nil)
	}
	return account, nil
}

func (h *CompleteCheckoutHandler) setPayer(ctx context.Context, r domain.RepoRegistry, account *domain.BillingAccount, payer domain.Payer, now time.Time) error {
	if account.Payer().ID() == payer.ID() && account.PayerMissingSince() == nil {
		return nil
	}
	account.ChangePayer(payer, now, payer.ID())
	return r.BillingAccount().Save(ctx, account)
}

// subscribe fetches the new subscription's payments up front, so the duplicate decision can be made
// inside the transaction, under the account lock.
func (h *CompleteCheckoutHandler) subscribe(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	snap, err := h.provider.GetSubscriptionSnapshot(ctx, s.ProviderSubscriptionID)
	if err != nil {
		return err
	}
	payments, err := h.provider.PeriodPayments(ctx, s.ProviderSubscriptionID, snap.CurrentPeriodStart(), snap.CurrentPeriodEnd())
	if err != nil {
		return err
	}

	now, actor := time.Now(), h.settings.ServiceAccountUserID
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
			return err
		}
		account, err := h.lockAccount(ctx, r, s.WorkspaceID)
		if err != nil {
			return err
		}
		if existing, err := r.Subscription().GetByProviderSubscriptionID(ctx, s.ProviderSubscriptionID, true); err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetByProviderPriceID(ctx, snap.ProviderPriceID())
		if err != nil {
			return err
		}
		cur, err := r.Subscription().GetCurrentByWorkspace(ctx, s.WorkspaceID, false)
		if err != nil {
			return err
		}
		id := uuid.Must(uuid.NewV7())
		if cur != nil {
			sub, err := domain.NewDuplicateSubscription(id, s.WorkspaceID, price, snap, payments, now, actor)
			if err != nil {
				return err
			}
			return r.Subscription().Save(ctx, sub)
		}
		sub, err := domain.NewSubscription(id, s.WorkspaceID, price, snap, now, actor)
		if err != nil {
			return err
		}
		if err := h.setPayer(ctx, r, account, s.Payer, now); err != nil {
			return err
		}
		return r.Subscription().Save(ctx, sub)
	})
}

// lifetime replaces a recurring subscription (refunding its unused time) or, over a lifetime one, is a duplicate.
// The replaced subscription's payments are fetched before the transaction; if the current subscription
// changed by then, the error makes the provider redeliver the event.
func (h *CompleteCheckoutHandler) lifetime(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	before, err := h.subs.GetCurrentByWorkspace(ctx, s.WorkspaceID, false)
	if err != nil {
		return err
	}
	var replaced []domain.PaymentInfo
	if before != nil && before.Interval() != domain.IntervalLifetime {
		if replaced, err = h.provider.PeriodPayments(ctx, *before.ProviderSubscriptionID(), *before.CurrentPeriodStart(), *before.CurrentPeriodEnd()); err != nil {
			return err
		}
	}

	now, actor := time.Now(), h.settings.ServiceAccountUserID
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
			return err
		}
		account, err := h.lockAccount(ctx, r, s.WorkspaceID)
		if err != nil {
			return err
		}
		if existing, err := r.Subscription().GetByProviderPaymentID(ctx, s.Payment.PaymentRef(), true); err != nil || existing != nil {
			return err
		}
		price, err := r.Price().GetByID(ctx, s.PriceID, false)
		if err != nil {
			return err
		}
		cur, err := r.Subscription().GetCurrentByWorkspace(ctx, s.WorkspaceID, true)
		if err != nil {
			return err
		}
		id := uuid.Must(uuid.NewV7())
		if cur != nil && cur.Interval() == domain.IntervalLifetime {
			sub, err := domain.NewDuplicateLifetimeSubscription(id, s.WorkspaceID, price, s.Payment, now, actor)
			if err != nil {
				return err
			}
			return r.Subscription().Save(ctx, sub)
		}
		if cur != nil {
			if before == nil || before.ID() != cur.ID() {
				return errs.NewInternal("current subscription changed while completing the lifetime checkout", nil)
			}
			policy, err := r.BillingPolicy().Get(ctx, false)
			if err != nil {
				return err
			}
			quote, err := cur.QuoteCancelNow(now, policy, replaced)
			if err != nil {
				return err
			}
			if _, err := cur.CancelNow(now, domain.RefundReasonReplacedByLifetime, quote, actor); err != nil {
				return err
			}
			if err := r.Subscription().Save(ctx, cur); err != nil {
				return err
			}
		}
		sub, err := domain.NewLifetimeSubscription(id, s.WorkspaceID, price, s.Payment.PaymentRef(), s.Payment.PaidAt(), now, actor)
		if err != nil {
			return err
		}
		if err := h.setPayer(ctx, r, account, s.Payer, now); err != nil {
			return err
		}
		return r.Subscription().Save(ctx, sub)
	})
}

func (h *CompleteCheckoutHandler) payerChange(ctx context.Context, cmd *CompleteCheckout, s CheckoutSessionResult) error {
	cur, err := h.subs.GetCurrentByWorkspace(ctx, s.WorkspaceID, false)
	if err != nil {
		return err
	}
	account, err := h.accounts.GetByWorkspace(ctx, s.WorkspaceID, false)
	if err != nil || account == nil {
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
		if first, err := r.ProviderEventLog().MarkProcessed(ctx, cmd.EventID, cmd.EventType); err != nil || !first {
			return err
		}
		locked, err := h.lockAccount(ctx, r, s.WorkspaceID)
		if err != nil {
			return err
		}
		if err := h.setPayer(ctx, r, locked, s.Payer, now); err != nil {
			return err
		}
		if !resume {
			return nil
		}
		sub, err := r.Subscription().GetByID(ctx, cur.ID(), true)
		if err != nil {
			return err
		}
		sub.ResumeCancel(now, s.Payer.ID())
		return r.Subscription().Save(ctx, sub)
	})
	return err
}
