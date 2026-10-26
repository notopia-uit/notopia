package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

// Notifier publishes best-effort notifications after commit. Implementations log errors and never fail.
type Notifier interface {
	SubscriptionChanged(ctx context.Context, workspaceID uuid.UUID, plan domain.PlanTier, status domain.Status)
}

func notifySubscription(ctx context.Context, n Notifier, s *domain.Subscription) {
	if s == nil {
		return
	}
	n.SubscriptionChanged(ctx, s.WorkspaceID(), s.Plan(), s.Status())
}
