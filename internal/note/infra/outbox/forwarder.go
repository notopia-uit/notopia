package outbox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgrepo"
	"github.com/notopia-uit/notopia/pkg/metadata"
)

// NOTE: May not need to use correlation id if we use otel, but we have to set msg context

type ForwarderPublisher struct {
	workspaceIDKey string
	aggregateIDKey string
	userIDKey      string
	publisher      message.Publisher
	jsonMarshaler  *cqrs.JSONMarshaler
}

var _ pgrepo.Publisher = (*ForwarderPublisher)(nil)

// TODO: create topic, metadata...
// TODO: If otel already, we can remove correlation ID
func (p *ForwarderPublisher) PublishWorkspaceItem(ctx context.Context, event domain.Event, workspaceID uuid.UUID) error {
	slog.DebugContext(ctx, "Publishing workspace item event to forwarder publisher", slog.String("event_type", fmt.Sprintf("%T", event)), slog.String("workspace_id", workspaceID.String()), slog.String("aggregate_id", event.GetAggregateID().String()), slog.String("user_id", event.GetUserID()))
	msg, err := p.jsonMarshaler.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for forwarder publisher: %w", err)
	}
	msg.SetContext(ctx)
	middleware.SetCorrelationID(msg.UUID, msg)
	msg.Metadata.Set(p.workspaceIDKey, workspaceID.String())
	msg.Metadata.Set(p.aggregateIDKey, event.GetAggregateID().String())
	msg.Metadata.Set(p.userIDKey, event.GetUserID())
	topic, err := component.DomainEventToTopic(event)
	if err != nil {
		return fmt.Errorf("failed to get forwarder event bus topic: %w", err)
	}
	if err := p.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish forwarder SQL event to event bus: %w", err)
	}
	slog.InfoContext(ctx, "Published workspace item event to forwarder publisher", slog.String("event_type", fmt.Sprintf("%T", event)), slog.String("workspace_id", workspaceID.String()), slog.String("aggregate_id", event.GetAggregateID().String()), slog.String("user_id", event.GetUserID()))
	return nil
}

func (p *ForwarderPublisher) Publish(ctx context.Context, event domain.Event) error {
	slog.DebugContext(ctx, "Publishing event to forwarder publisher", slog.String("event_type", fmt.Sprintf("%T", event)), slog.String("aggregate_id", event.GetAggregateID().String()), slog.String("user_id", event.GetUserID()))
	msg, err := p.jsonMarshaler.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for forwarder publisher: %w", err)
	}
	msg.SetContext(ctx)
	middleware.SetCorrelationID(msg.UUID, msg)
	msg.Metadata.Set(p.aggregateIDKey, event.GetAggregateID().String())
	msg.Metadata.Set(p.userIDKey, event.GetUserID())
	topic, err := component.DomainEventToTopic(event)
	if err != nil {
		return fmt.Errorf("failed to get forwarder event bus topic: %w", err)
	}
	if err := p.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish forwarder SQL event to event bus: %w", err)
	}
	slog.InfoContext(ctx, "Published event to forwarder publisher", slog.String("event_type", fmt.Sprintf("%T", event)), slog.String("aggregate_id", event.GetAggregateID().String()), slog.String("user_id", event.GetUserID()))
	return nil
}

type FromPersistenceToQSLForwarder struct {
	workspaceIDKey string
	aggregateIDKey string
	userIDKey      string
	logger         watermill.LoggerAdapter
	schemaAdapter  sql.SchemaAdapter
	serviceName    string
	jsonMarshaler  *cqrs.JSONMarshaler
}

var _ pgrepo.PublisherFactory = (*FromPersistenceToQSLForwarder)(nil)

func NewFromPersistenceToQSLForwarder(
	domainEventCfg config.DomainEvent,
	logger watermill.LoggerAdapter,
	schemaAdapter sql.SchemaAdapter,
	serviceName metadata.ServiceName,
	jsonMarshaler *cqrs.JSONMarshaler,
) *FromPersistenceToQSLForwarder {
	return &FromPersistenceToQSLForwarder{
		workspaceIDKey: domainEventCfg.MessageWorkspaceIDKey,
		aggregateIDKey: domainEventCfg.MessageMetadataAggregateIDKey,
		userIDKey:      domainEventCfg.MessageMetadataUserIDKey,
		logger:         logger,
		schemaAdapter:  schemaAdapter,
		serviceName:    serviceName.String(),
		jsonMarshaler:  jsonMarshaler,
	}
}

var ProvideFromPersistenceToQSLForwarder = NewFromPersistenceToQSLForwarder

func (f *FromPersistenceToQSLForwarder) Create(
	ctx context.Context,
	pgxTx pgx.Tx,
) (pgrepo.Publisher, error) {
	slog.DebugContext(ctx, "Creating forwarder publisher for persistence to QSL forwarder")
	sqlPublisher, err := sql.NewPublisher(
		sql.TxFromPgx(pgxTx),
		sql.PublisherConfig{
			SchemaAdapter:        f.schemaAdapter,
			AutoInitializeSchema: false,
		},
		watermill.NopLogger{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL publisher: %w", err)
	}
	publisher := forwarder.NewPublisher(sqlPublisher, forwarder.PublisherConfig{})
	otelPublisher := wotel.NewNamedPublisherDecorator(
		f.serviceName,
		publisher,
	)
	slog.DebugContext(ctx, "Created forwarder publisher for persistence to QSL forwarder")
	return &ForwarderPublisher{
		workspaceIDKey: f.workspaceIDKey,
		aggregateIDKey: f.aggregateIDKey,
		userIDKey:      f.userIDKey,
		publisher:      otelPublisher,
		jsonMarshaler:  f.jsonMarshaler,
	}, nil
}
