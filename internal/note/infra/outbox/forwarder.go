package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgrepo"
)

type ForwarderPublisher struct {
	workspaceIDKey string
	aggregateIDKey string
	userIDKey      string
	publisher      message.Publisher
}

var _ pgrepo.Publisher = (*ForwarderPublisher)(nil)

// TODO: create topic, metadata...
func (p *ForwarderPublisher) PublishWorkspaceItem(ctx context.Context, event domain.Event, workspaceID uuid.UUID) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for forwarder publisher: %w", err)
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set(p.workspaceIDKey, workspaceID.String())
	msg.Metadata.Set(p.aggregateIDKey, event.GetAggregateID().String())
	msg.Metadata.Set(p.userIDKey, event.GetUserID())
	topic, ok := component.DomainEventToTopic(event)
	if !ok {
		return fmt.Errorf("failed to get forwader event bus topic for event type: %T", event)
	}
	if err := p.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish forwarder SQL event to event bus: %w", err)
	}
	return nil
}

func (p *ForwarderPublisher) Publish(ctx context.Context, event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for forwarder publisher: %w", err)
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	topic, ok := component.DomainEventToTopic(event)
	if !ok {
		return fmt.Errorf("failed to get forwader event bus topic for event type: %T", event)
	}
	if err := p.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish forwarder SQL event to event bus: %w", err)
	}
	return nil
}

type FromPersistenceToQSLForwarder struct {
	workspaceIDKey string
	aggregateIDKey string
	userIDKey      string
	logger         watermill.LoggerAdapter
}

var _ pgrepo.PublisherFactory = (*FromPersistenceToQSLForwarder)(nil)

func NewFromPersistenceToQSLForwarder(
	domainEventCfg config.DomainEvent,
	logger watermill.LoggerAdapter,
) *FromPersistenceToQSLForwarder {
	return &FromPersistenceToQSLForwarder{
		workspaceIDKey: domainEventCfg.MessageWorkspaceIDKey,
		aggregateIDKey: domainEventCfg.MessageMetadataAggregateIDKey,
		userIDKey:      domainEventCfg.MessageMetadataUserIDKey,
		logger:         logger,
	}
}

var ProvideFromPersistenceToQSLForwarder = NewFromPersistenceToQSLForwarder

func (f *FromPersistenceToQSLForwarder) Create(
	pgxTx pgx.Tx,
) (pgrepo.Publisher, error) {
	sqlPublisher, err := sql.NewPublisher(
		sql.TxFromPgx(pgxTx),
		sql.PublisherConfig{
			SchemaAdapter:        sql.DefaultPostgreSQLSchema{},
			AutoInitializeSchema: true,
		},
		watermill.NopLogger{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL publisher: %w", err)
	}
	publisher := forwarder.NewPublisher(sqlPublisher, forwarder.PublisherConfig{})
	return &ForwarderPublisher{
		workspaceIDKey: f.workspaceIDKey,
		aggregateIDKey: f.aggregateIDKey,
		userIDKey:      f.userIDKey,
		publisher:      publisher,
	}, nil
}
