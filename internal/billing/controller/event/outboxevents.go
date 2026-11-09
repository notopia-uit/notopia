package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
)

// subscriptionCanceledNowMessage and refundRequestedMessage decode the fields billing's own domain events
// carry on the outbox topics (domain.SubscriptionCanceledNowEvent, domain.RefundRequestedEvent).
type subscriptionCanceledNowMessage struct {
	AggregateID uuid.UUID
}

type refundRequestedMessage struct {
	AggregateID uuid.UUID
	RefundID    uuid.UUID
}

func (e *Event) subscriptionCanceledNow(ctx context.Context, ev *subscriptionCanceledNowMessage) error {
	return e.app.Cmds.CancelProviderSubscription.Handle(ctx, &app.CancelProviderSubscription{SubscriptionID: ev.AggregateID})
}

func (e *Event) refundRequested(ctx context.Context, ev *refundRequestedMessage) error {
	return e.app.Cmds.SubmitRefund.Handle(ctx, &app.SubmitRefund{SubscriptionID: ev.AggregateID, RefundID: ev.RefundID})
}
