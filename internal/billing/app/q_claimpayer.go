package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ClaimPayer struct {
	WorkspaceID uuid.UUID
	Payer       Payer
}

type ClaimPayerHandler struct {
	accounts domain.BillingAccountRepo
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewClaimPayerHandler(accounts domain.BillingAccountRepo, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *ClaimPayerHandler {
	return &ClaimPayerHandler{accounts: accounts, authz: authz, provider: provider, settings: settings}
}

var ProvideClaimPayerHandler = NewClaimPayerHandler

type ClaimPayerQuery commonhandler.Query[ClaimPayer, string]

var _ ClaimPayerQuery = (*ClaimPayerHandler)(nil)

func (h *ClaimPayerHandler) Handle(ctx context.Context, q *ClaimPayer) (string, error) {
	if err := requirePermission(ctx, h.authz, q.Payer.ID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	account, err := h.accounts.GetByWorkspace(ctx, q.WorkspaceID, false)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", errs.New(errs.CodeNoBillingAccount, "workspace has no billing account", nil)
	}
	payer, err := toDomainPayer(q.Payer)
	if err != nil {
		return "", err
	}
	if account.Payer().ID() == payer.ID() && account.PayerMissingSince() == nil {
		return "", errs.New(errs.CodeAlreadyPayer, "you are already the payer", nil)
	}
	return h.provider.CreateCheckoutSession(ctx, CheckoutInput{
		Purpose:         CheckoutPurposePayerChange,
		WorkspaceID:     q.WorkspaceID,
		CustomerID:      account.ProviderCustomerID(),
		Payer:           payer,
		PriceID:         uuid.Nil,
		ProviderPriceID: "",
		SuccessURL:      billingPageURL(h.settings, q.WorkspaceID, "payer"),
		CancelURL:       billingPageURL(h.settings, q.WorkspaceID, "cancel"),
	})
}
