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
	provider PaymentProvider,
	settings *Settings,
) *EnsurePortalConfigurationHandler {
	return &EnsurePortalConfigurationHandler{catalog: newCatalog(uow, provider, settings)}
}

var ProvideEnsurePortalConfigurationHandler = NewEnsurePortalConfigurationHandler

type EnsurePortalConfigurationCmd commonhandler.Cmd[EnsurePortalConfiguration]

var _ EnsurePortalConfigurationCmd = (*EnsurePortalConfigurationHandler)(nil)

// Handle creates the portal configuration only when billing_policy has none;
// CreatePrice/ArchivePrice already refresh it when the price set changes.
func (h *EnsurePortalConfigurationHandler) Handle(ctx context.Context, cmd *EnsurePortalConfiguration) error {
	var policy *domain.BillingPolicy
	if err := h.catalog.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		policy, err = r.BillingPolicy().Get(ctx, false)
		return err
	}); err != nil {
		return err
	}
	if policy.PortalConfigurationID() != nil {
		return nil
	}
	return h.catalog.refreshPortal(ctx, cmd.UserID)
}
