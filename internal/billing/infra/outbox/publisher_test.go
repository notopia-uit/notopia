package outbox

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPublisher() *Publisher {
	return &Publisher{
		cfg: &config.DomainEvent{
			MessageMetadataAggregateIDKey:  "aggregate_id",
			MessageMetadataUserIDKey:       "user_id",
			MessageMetadataPartitionKeyKey: "partition_key",
		},
		publisher: nil,
		marshaler: cqrs.JSONMarshaler{},
	}
}

func TestStateChangedIsMappedToTheIntegrationContract(t *testing.T) {
	ws, subID := uuid.New(), uuid.New()
	event := &domain.SubscriptionStateChangedEvent{
		BaseEvent:   domain.NewBaseEvent(subID.String(), "u"),
		WorkspaceID: ws,
		Plan:        domain.PlanTierPlus,
		Status:      domain.SubscriptionStatusPastDue,
	}

	msg, err := newTestPublisher().toMessage(context.Background(), event)

	require.NoError(t, err)
	assert.JSONEq(t, `{"workspaceId":"`+ws.String()+`","plan":"plus","status":"past_due"}`, string(msg.Payload))
	assert.Equal(t, event.ID.String(), msg.UUID)
	assert.Equal(t, ws.String(), msg.Metadata.Get("partition_key"))
	assert.Equal(t, subID.String(), msg.Metadata.Get("aggregate_id"))
}

func TestDomainEventKeepsIDsConsumersDecode(t *testing.T) {
	subID, refundID := uuid.New(), uuid.New()
	event := &domain.RefundRequestedEvent{
		BaseEvent: domain.NewBaseEvent(subID.String(), "u"),
		RefundID:  refundID,
	}

	msg, err := newTestPublisher().toMessage(context.Background(), event)

	require.NoError(t, err)
	assert.Contains(t, string(msg.Payload), `"AggregateID":"`+subID.String()+`"`)
	assert.Contains(t, string(msg.Payload), `"RefundID":"`+refundID.String()+`"`)
	assert.Equal(t, subID.String(), msg.Metadata.Get("partition_key"))
}
