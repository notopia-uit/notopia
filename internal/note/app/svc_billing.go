package app

import (
	"context"

	"github.com/google/uuid"
)

type BillingSvc interface {
	GetEntitlements(ctx context.Context, workspaceID uuid.UUID) (Entitlements, error)
	BatchGetEntitlements(ctx context.Context, workspaceIDs []uuid.UUID) (map[uuid.UUID]Entitlements, error)
	GetPayer(ctx context.Context, workspaceID uuid.UUID) (payerID *string, requiresPayer bool, err error)
	ListPlans(ctx context.Context) ([]PlanEntitlements, error)
}
