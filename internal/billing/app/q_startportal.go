package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type StartPortal struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type StartPortalHandler struct {
	accounts domain.BillingAccountRepo
	policies domain.BillingPolicyRepo
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewStartPortalHandler(accounts domain.BillingAccountRepo, policies domain.BillingPolicyRepo, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *StartPortalHandler {
	return &StartPortalHandler{accounts: accounts, policies: policies, authz: authz, provider: provider, settings: settings}
}

var ProvideStartPortalHandler = NewStartPortalHandler

type StartPortalQuery commonhandler.Query[StartPortal, string]

var _ StartPortalQuery = (*StartPortalHandler)(nil)

func (h *StartPortalHandler) Handle(ctx context.Context, q *StartPortal) (string, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	account, err := h.accounts.GetByWorkspace(ctx, q.WorkspaceID, false)
	if err != nil {
		return "", err
	}
	policy, err := h.policies.Get(ctx, false)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", errs.New(errs.CodeNoBillingAccount, "workspace has no billing account", nil)
	}
	if policy.PortalConfigurationID() == nil {
		return "", errs.NewInternal("portal configuration missing: run just seed-billing", nil)
	}
	return h.provider.CreatePortalSession(ctx, account.ProviderCustomerID(), *policy.PortalConfigurationID(),
		fmt.Sprintf("%s/workspace/%s/settings/billing?portal=return", h.settings.WebBaseURL, q.WorkspaceID))
}
