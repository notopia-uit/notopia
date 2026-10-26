package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event interface {
	GetID() uuid.UUID
	GetAggregateID() string
	GetUserID() string
}

type BaseEvent struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	AggregateID string
	UserID      string
}

func NewBaseEvent(aggregateID, userID string) BaseEvent {
	return BaseEvent{ID: uuid.Must(uuid.NewV7()), OccurredAt: time.Now(), AggregateID: aggregateID, UserID: userID}
}

func (e *BaseEvent) GetID() uuid.UUID       { return e.ID }
func (e *BaseEvent) GetAggregateID() string { return e.AggregateID }
func (e *BaseEvent) GetUserID() string      { return e.UserID }

type eventRecorder struct{ events []Event }

func (r *eventRecorder) record(e Event) { r.events = append(r.events, e) }

func (r *eventRecorder) PopEvents() []Event {
	events := r.events
	r.events = nil
	return events
}

type (
	PlanUpdatedEvent struct {
		BaseEvent
		Features map[FeatureKey]bool
		Limits   map[LimitKey]int64
	}
	BillingPolicyUpdatedEvent       struct{ BaseEvent }
	PortalConfigurationChangedEvent struct {
		BaseEvent
		ConfigurationID string
	}
	PriceCreatedEvent  struct{ BaseEvent }
	PriceArchivedEvent struct{ BaseEvent }

	BillingAccountOpenedEvent struct{ BaseEvent }
	PayerChangedEvent         struct {
		BaseEvent
		PayerID string
	}
	PayerMissingEvent          struct{ BaseEvent }
	BillingContactUpdatedEvent struct{ BaseEvent }

	SubscriptionCreatedEvent struct {
		BaseEvent
		Plan     PlanTier
		Interval Interval
	}
	SubscriptionActivatedEvent   struct{ BaseEvent }
	SubscriptionRenewedEvent     struct{ BaseEvent }
	SubscriptionPastDueEvent     struct{ BaseEvent }
	SubscriptionPlanChangedEvent struct {
		BaseEvent
		From, To PlanTier
	}
	SubscriptionCancelScheduledEvent struct{ BaseEvent }
	SubscriptionCancelResumedEvent   struct{ BaseEvent }
	SubscriptionCanceledNowEvent     struct {
		BaseEvent
		Reason RefundReason
	}
	SubscriptionEndedEvent struct{ BaseEvent }
	RefundRequestedEvent   struct {
		BaseEvent
		RefundID uuid.UUID
		Amount   Money
	}
	RefundSucceededEvent struct {
		BaseEvent
		RefundID uuid.UUID
	}
	RefundFailedEvent struct {
		BaseEvent
		RefundID uuid.UUID
	}
)
