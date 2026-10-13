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
	uow      domain.UnitOfWork
	authz    AuthorizationSvc
	provider PaymentProvider
	settings *Settings
}

func NewStartPortalHandler(uow domain.UnitOfWork, authz AuthorizationSvc, provider PaymentProvider, settings *Settings) *StartPortalHandler {
	return &StartPortalHandler{uow: uow, authz: authz, provider: provider, settings: settings}
}

var ProvideStartPortalHandler = NewStartPortalHandler

type StartPortalQuery commonhandler.Query[StartPortal, string]

var _ StartPortalQuery = (*StartPortalHandler)(nil)

func (h *StartPortalHandler) Handle(ctx context.Context, q *StartPortal) (string, error) {
	if err := requirePermission(ctx, h.authz, q.UserID, q.WorkspaceID, WorkspacePermissionBillingManage); err != nil {
		return "", err
	}
	var account *domain.BillingAccount
	var policy *domain.BillingPolicy
	if err := h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if account, err = r.BillingAccount().GetByWorkspace(ctx, q.WorkspaceID, false); err != nil {
			return err
		}
		policy, err = r.BillingPolicy().Get(ctx, false)
		return err
	}); err != nil {
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
