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
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewStartCheckoutHandler(uow domain.UnitOfWork, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *StartCheckoutHandler {
	return &StartCheckoutHandler{uow: uow, authz: authz, provider: provider, settings: settings}
}

var ProvideStartCheckoutHandler = NewStartCheckoutHandler

type StartCheckoutQuery commonhandler.Query[StartCheckout, string]

var _ StartCheckoutQuery = (*StartCheckoutHandler)(nil)

func billingPageURL(settings *Settings, ws uuid.UUID, result string) string {
	return fmt.Sprintf("%s/workspace/%s/settings/billing?checkout=%s", settings.WebBaseURL, ws, result)
}

// ensureAccount opens the billing account (and Stripe customer) on first use.
func ensureAccount(ctx context.Context, uow domain.UnitOfWork, provider PaymentProvider, ws uuid.UUID, payer domain.Payer) (*domain.BillingAccount, error) {
	var account *domain.BillingAccount
	if err := uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		account, err = r.BillingAccount().GetByWorkspace(ctx, ws, false)
		return err
	}); err != nil || account != nil {
		return account, err
	}
	customerID, err := provider.CreateCustomer(ctx, ws, payer)
	if err != nil {
		return nil, err
	}
	err = uow.Execute(ctx, func(r domain.RepoRegistry) error {
		existing, err := r.BillingAccount().GetByWorkspace(ctx, ws, true)
		if err != nil || existing != nil {
			account = existing
			return err
		}
		account = domain.OpenBillingAccount(ws, customerID, payer, time.Now(), payer.ID)
		return r.BillingAccount().Save(ctx, account)
	})
	return account, err
}

func (h *StartCheckoutHandler) Handle(ctx context.Context, q *StartCheckout) (string, error) {
	if err := requirePermission(ctx, h.authz, q.Payer.ID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	var price *domain.Price
	var cur *domain.Subscription
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if price, err = r.Price().GetByID(ctx, q.PriceID, false); err != nil {
			return err
		}
		cur, err = r.Subscription().GetCurrentByWorkspace(ctx, q.WorkspaceID, false)
		return err
	}); err != nil {
		return "", err
	}
	if !price.Active() {
		return "", errs.New(errs.CodePriceNotActive, "price is archived", nil)
	}
	lifetimeUpgrade := price.Interval() == domain.IntervalLifetime && cur != nil && cur.Interval() != domain.IntervalLifetime
	if cur != nil && !lifetimeUpgrade {
		return "", errs.New(errs.CodeSubscriptionAlreadyActive, "workspace already has a subscription; use the portal to change plan", nil)
	}
	payer := toDomainPayer(q.Payer)
	account, err := ensureAccount(ctx, h.uow, h.provider, q.WorkspaceID, payer)
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
