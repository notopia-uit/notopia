package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

type BillingReadModel interface {
	CurrentSubscription(ctx context.Context, workspaceID uuid.UUID) (*domain.Subscription, error) // nil, nil when none
	CurrentSubscriptions(ctx context.Context, workspaceIDs []uuid.UUID) ([]*domain.Subscription, error)
	Account(ctx context.Context, workspaceID uuid.UUID) (*domain.BillingAccount, error) // nil, nil when none
	Plans(ctx context.Context) (map[domain.PlanTier]*domain.Plan, error)
	Policy(ctx context.Context) (*domain.BillingPolicy, error)
	ActivePrices(ctx context.Context) ([]*domain.Price, error)
	Prices(ctx context.Context) ([]*domain.Price, error)
}
