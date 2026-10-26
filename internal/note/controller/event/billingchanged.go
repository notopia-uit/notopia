package event

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func toAppPlanTierFromShare(p share.SubscriptionChangedEventPlan) (app.PlanTier, error) {
	switch p {
	case share.SubscriptionChangedEventPlanFree:
		return app.PlanTierFree, nil
	case share.SubscriptionChangedEventPlanPlus:
		return app.PlanTierPlus, nil
	case share.SubscriptionChangedEventPlanPro:
		return app.PlanTierPro, nil
	default:
		return app.PlanTierUnspecified, fmt.Errorf("invalid plan in SubscriptionChangedEvent: %q", p)
	}
}

func (e *Event) subscriptionChangedHandler(msg *message.Message) error {
	var ev share.SubscriptionChangedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal SubscriptionChangedEvent: %w", err)
	}
	plan, err := toAppPlanTierFromShare(ev.Plan)
	if err != nil {
		slog.WarnContext(msg.Context(), "dropping billing notification", slog.Any("error", err))
		return nil
	}
	return e.app.Events.BillingChanged.Handle(
		msg.Context(),
		&app.BillingChanged{WorkspaceID: ev.WorkspaceId, Plan: plan},
	)
}
