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
	uow      domain.UnitOfWork
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewClaimPayerHandler(uow domain.UnitOfWork, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *ClaimPayerHandler {
	return &ClaimPayerHandler{uow: uow, authz: authz, provider: provider, settings: settings}
}

var ProvideClaimPayerHandler = NewClaimPayerHandler

type ClaimPayerQuery commonhandler.Query[ClaimPayer, string]

var _ ClaimPayerQuery = (*ClaimPayerHandler)(nil)

func (h *ClaimPayerHandler) Handle(ctx context.Context, q *ClaimPayer) (string, error) {
	if err := requirePermission(ctx, h.authz, q.Payer.ID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	var account *domain.BillingAccount
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		account, err = r.BillingAccount().GetByWorkspace(ctx, q.WorkspaceID, false)
		return err
	}); err != nil {
		return "", err
	}
	if account == nil {
		return "", errs.New(errs.CodeNoBillingAccount, "workspace has no billing account", nil)
	}
	if account.Payer().ID == q.Payer.ID && account.PayerMissingSince() == nil {
		return "", errs.New(errs.CodeAlreadyPayer, "you are already the payer", nil)
	}
	return h.provider.CreateCheckoutSession(ctx, CheckoutInput{
		Purpose:         CheckoutPurposePayerChange,
		WorkspaceID:     q.WorkspaceID,
		CustomerID:      account.ProviderCustomerID(),
		Payer:           toDomainPayer(q.Payer),
		PriceID:         uuid.Nil,
		ProviderPriceID: "",
		SuccessURL:      billingPageURL(h.settings, q.WorkspaceID, "payer"),
		CancelURL:       billingPageURL(h.settings, q.WorkspaceID, "cancel"),
	})
}
