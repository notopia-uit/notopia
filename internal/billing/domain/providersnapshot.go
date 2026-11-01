package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type ProviderSnapshot struct {
	providerSubscriptionID string
	providerPriceID        string
	workspaceID            uuid.UUID
	status                 SubscriptionStatus
	currentPeriodStart     time.Time
	currentPeriodEnd       time.Time
	cancelAtPeriodEnd      bool
	canceledAt             *time.Time
}

type ProviderSnapshotParams struct {
	ProviderSubscriptionID string
	ProviderPriceID        string
	WorkspaceID            uuid.UUID
	Status                 SubscriptionStatus
	CurrentPeriodStart     time.Time
	CurrentPeriodEnd       time.Time
	CancelAtPeriodEnd      bool
	CanceledAt             *time.Time
}

func NewProviderSnapshot(p ProviderSnapshotParams) (ProviderSnapshot, error) {
	if p.ProviderSubscriptionID == "" {
		return ProviderSnapshot{}, errs.NewInvalid("provider snapshot needs a subscription id")
	}
	if !p.Status.IsValid() {
		return ProviderSnapshot{}, errs.NewInvalid("provider snapshot has an invalid status")
	}
	if p.CurrentPeriodEnd.Before(p.CurrentPeriodStart) {
		return ProviderSnapshot{}, errs.NewInvalid("provider snapshot period ends before it starts")
	}
	var canceledAt *time.Time
	if p.CanceledAt != nil {
		canceledAt = new(*p.CanceledAt)
	}
	return ProviderSnapshot{
		providerSubscriptionID: p.ProviderSubscriptionID,
		providerPriceID:        p.ProviderPriceID,
		workspaceID:            p.WorkspaceID,
		status:                 p.Status,
		currentPeriodStart:     p.CurrentPeriodStart,
		currentPeriodEnd:       p.CurrentPeriodEnd,
		cancelAtPeriodEnd:      p.CancelAtPeriodEnd,
		canceledAt:             canceledAt,
	}, nil
}

func (s ProviderSnapshot) ProviderSubscriptionID() string { return s.providerSubscriptionID }
func (s ProviderSnapshot) ProviderPriceID() string        { return s.providerPriceID }
func (s ProviderSnapshot) WorkspaceID() uuid.UUID         { return s.workspaceID }
func (s ProviderSnapshot) Status() SubscriptionStatus     { return s.status }
func (s ProviderSnapshot) CurrentPeriodStart() time.Time  { return s.currentPeriodStart }
func (s ProviderSnapshot) CurrentPeriodEnd() time.Time    { return s.currentPeriodEnd }
func (s ProviderSnapshot) CancelAtPeriodEnd() bool        { return s.cancelAtPeriodEnd }

func (s ProviderSnapshot) CanceledAt() *time.Time {
	if s.canceledAt == nil {
		return nil
	}
	return new(*s.canceledAt)
}
