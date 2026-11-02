package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type StartCheckout struct {
	WorkspaceID uuid.UUID
	PriceID     uuid.UUID
	Payer       Payer
}

type StartCheckoutHandler struct {
	uow      domain.UnitOfWork
	accounts domain.BillingAccountRepo
	prices   domain.PriceRepo
	subs     domain.SubscriptionRepo
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewStartCheckoutHandler(
	uow domain.UnitOfWork,
	accounts domain.BillingAccountRepo,
	prices domain.PriceRepo,
	subs domain.SubscriptionRepo,
	authz AuthorizationSvc,
	provider PaymentProvider,
	settings *Settings,
) *StartCheckoutHandler {
	return &StartCheckoutHandler{uow: uow, accounts: accounts, prices: prices, subs: subs, authz: authz, provider: provider, settings: settings}
}

var ProvideStartCheckoutHandler = NewStartCheckoutHandler

type StartCheckoutQuery commonhandler.Query[StartCheckout, string]

var _ StartCheckoutQuery = (*StartCheckoutHandler)(nil)

func billingPageURL(settings *Settings, ws uuid.UUID, result string) string {
	return fmt.Sprintf("%s/workspace/%s/settings/billing?checkout=%s", settings.WebBaseURL, ws, result)
}

// ensureAccount opens the billing account (and Stripe customer) on first use.
func (h *StartCheckoutHandler) ensureAccount(ctx context.Context, ws uuid.UUID, payer domain.Payer) (*domain.BillingAccount, error) {
	account, err := h.accounts.GetByWorkspace(ctx, ws, false)
	if err != nil || account != nil {
		return account, err
	}
	customerID, err := h.provider.CreateCustomer(ctx, ws, payer)
	if err != nil {
		return nil, err
	}
	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		existing, err := r.BillingAccount().GetByWorkspace(ctx, ws, true)
		if err != nil || existing != nil {
			account = existing
			return err
		}
		account = domain.OpenBillingAccount(ws, customerID, payer, time.Now(), payer.ID())
		return r.BillingAccount().Save(ctx, account)
	})
	return account, err
}

func (h *StartCheckoutHandler) Handle(ctx context.Context, q *StartCheckout) (string, error) {
	if err := requirePermission(ctx, h.authz, q.Payer.ID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	price, err := h.prices.GetByID(ctx, q.PriceID, false)
	if err != nil {
		return "", err
	}
	cur, err := h.subs.GetCurrentByWorkspace(ctx, q.WorkspaceID, false)
	if err != nil {
		return "", err
	}
	if !price.Active() {
		return "", errs.New(errs.CodePriceNotActive, "price is archived", nil)
	}
	lifetimeUpgrade := price.Interval() == domain.IntervalLifetime && cur != nil && cur.Interval() != domain.IntervalLifetime
	if cur != nil && !lifetimeUpgrade {
		return "", errs.New(errs.CodeSubscriptionAlreadyActive, "workspace already has a subscription; use the portal to change plan", nil)
	}
	payer, err := toDomainPayer(q.Payer)
	if err != nil {
		return "", err
	}
	account, err := h.ensureAccount(ctx, q.WorkspaceID, payer)
	if err != nil {
		return "", err
	}
	purpose := CheckoutPurposeSubscribe
	if price.Interval() == domain.IntervalLifetime {
		purpose = CheckoutPurposeLifetime
	}
	return h.provider.CreateCheckoutSession(ctx, CheckoutInput{
		Purpose: purpose, WorkspaceID: q.WorkspaceID, CustomerID: account.ProviderCustomerID(), Payer: payer,
		PriceID: price.ID(), ProviderPriceID: price.ProviderPriceID(),
		SuccessURL: billingPageURL(h.settings, q.WorkspaceID, "success"), CancelURL: billingPageURL(h.settings, q.WorkspaceID, "cancel"),
	})
}
