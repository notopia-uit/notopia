package outbox

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func toShareSubscriptionChanged(e *domain.SubscriptionStateChangedEvent) (*share.SubscriptionChangedEvent, error) {
	plan, err := toSharePlan(e.Plan)
	if err != nil {
		return nil, err
	}
	status, err := toShareStatus(e.Status)
	if err != nil {
		return nil, err
	}
	return &share.SubscriptionChangedEvent{
		WorkspaceId: e.WorkspaceID,
		Plan:        plan,
		Status:      status,
	}, nil
}

func toSharePlan(t domain.PlanTier) (share.SubscriptionChangedEventPlan, error) {
	switch t {
	case domain.PlanTierFree:
		return share.SubscriptionChangedEventPlanFree, nil
	case domain.PlanTierPlus:
		return share.SubscriptionChangedEventPlanPlus, nil
	case domain.PlanTierPro:
		return share.SubscriptionChangedEventPlanPro, nil
	case domain.PlanTierUnspecified:
	}
	return "", fmt.Errorf("invalid plan tier: %v", t)
}

func toShareStatus(s domain.SubscriptionStatus) (string, error) {
	if !s.IsValid() {
		return "", fmt.Errorf("invalid subscription status: %v", s)
	}
	return s.String(), nil
}
