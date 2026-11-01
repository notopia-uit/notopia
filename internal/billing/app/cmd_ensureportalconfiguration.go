package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type EnsurePortalConfiguration struct {
	UserID string
}

type EnsurePortalConfigurationHandler struct{ catalog *catalog }

func NewEnsurePortalConfigurationHandler(
	uow domain.UnitOfWork,
	prices domain.PriceRepo,
	policies domain.BillingPolicyRepo,
	provider PaymentProvider,
	settings *Settings,
) *EnsurePortalConfigurationHandler {
	return &EnsurePortalConfigurationHandler{catalog: newCatalog(uow, prices, policies, provider, settings)}
}

var ProvideEnsurePortalConfigurationHandler = NewEnsurePortalConfigurationHandler

type EnsurePortalConfigurationCmd commonhandler.Cmd[EnsurePortalConfiguration]

var _ EnsurePortalConfigurationCmd = (*EnsurePortalConfigurationHandler)(nil)

// Handle creates the portal configuration only when billing_policy has none;
// CreatePrice/ArchivePrice already refresh it when the price set changes.
func (h *EnsurePortalConfigurationHandler) Handle(ctx context.Context, cmd *EnsurePortalConfiguration) error {
	policy, err := h.catalog.policies.Get(ctx, false)
	if err != nil {
		return err
	}
	if policy.PortalConfigurationID() != nil {
		return nil
	}
	return h.catalog.refreshPortal(ctx, cmd.UserID)
}
