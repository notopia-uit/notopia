package domain

import (
	"time"

	"github.com/google/uuid"
)

type Entitlements struct {
	WorkspaceID uuid.UUID
	Plan        PlanTier
	Status      Status
	AccessUntil *time.Time
	Features    map[FeatureKey]bool
	Limits      map[LimitKey]int64
}

func ResolveEntitlements(workspaceID uuid.UUID, sub *Subscription, plans map[PlanTier]*Plan, policy *BillingPolicy, now time.Time) Entitlements {
	tier, status := PlanTierFree, StatusNone
	var until *time.Time
	if sub != nil {
		status = sub.Status()
		if sub.GrantsAccess(now, policy) {
			tier, until = sub.Plan(), sub.AccessUntil(policy)
		}
	}
	plan := plans[tier]
	return Entitlements{WorkspaceID: workspaceID, Plan: tier, Status: status, AccessUntil: until, Features: plan.Features(), Limits: plan.Limits()}
}
