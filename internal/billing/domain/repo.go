package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Subscription, error)
	GetCurrentByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*Subscription, error) // nil, nil when none
	GetByProviderSubscriptionID(ctx context.Context, id string, forUpdate bool) (*Subscription, error)       // nil, nil when none
	GetByProviderPaymentID(ctx context.Context, id string, forUpdate bool) (*Subscription, error)            // nil, nil when none
	GetByRefundID(ctx context.Context, refundID uuid.UUID, forUpdate bool) (*Subscription, error)            // nil, nil when none
	ListCurrentByWorkspaces(ctx context.Context, workspaceIDs []uuid.UUID) ([]*Subscription, error)
	ListNeedingReconcile(ctx context.Context, now time.Time, limit int) ([]*Subscription, error)
	ListIDsWithUnsubmittedRefunds(ctx context.Context, limit int) ([]uuid.UUID, error)
	Save(ctx context.Context, s *Subscription) error
}

type BillingAccountRepo interface {
	GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*BillingAccount, error) // nil, nil when none
	ListByPayer(ctx context.Context, payerID string) ([]*BillingAccount, error)
	Save(ctx context.Context, a *BillingAccount) error
}

type PriceRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Price, error)
	GetByProviderPriceID(ctx context.Context, id string) (*Price, error)
	GetActive(ctx context.Context, plan PlanTier, interval Interval, currency string, forUpdate bool) (*Price, error) // nil, nil when none
	ListActive(ctx context.Context) ([]*Price, error)
	List(ctx context.Context) ([]*Price, error)
	Save(ctx context.Context, p *Price) error
}

type PlanRepo interface {
	Get(ctx context.Context, tier PlanTier, forUpdate bool) (*Plan, error)
	ListMap(ctx context.Context) (map[PlanTier]*Plan, error)
	Save(ctx context.Context, p *Plan) error
}

type BillingPolicyRepo interface {
	Get(ctx context.Context, forUpdate bool) (*BillingPolicy, error)
	Save(ctx context.Context, p *BillingPolicy) error
}

// ProviderEventLog records processed payment-provider events in the same transaction as their effect.
type ProviderEventLog interface {
	MarkProcessed(ctx context.Context, eventID, eventType string) (firstTime bool, err error)
	PruneBefore(ctx context.Context, before time.Time) (int64, error)
}

type RepoRegistry interface {
	Subscription() SubscriptionRepo
	BillingAccount() BillingAccountRepo
	Price() PriceRepo
	Plan() PlanRepo
	BillingPolicy() BillingPolicyRepo
	ProviderEventLog() ProviderEventLog
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(r RepoRegistry) error) error
}
