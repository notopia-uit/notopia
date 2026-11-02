package outbox

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgrepo"
	"github.com/notopia-uit/notopia/pkg/metadata"
)

type PublisherFactory struct {
	cfg           *config.DomainEvent
	schemaAdapter sql.SchemaAdapter
	marshaler     cqrs.JSONMarshaler
	serviceName   string
}

var _ pgrepo.PublisherFactory = (*PublisherFactory)(nil)

func NewPublisherFactory(
	cfg *config.DomainEvent,
	schemaAdapter sql.SchemaAdapter,
	serviceName metadata.ServiceName,
) *PublisherFactory {
	return &PublisherFactory{
		cfg:           cfg,
		schemaAdapter: schemaAdapter,
		marshaler:     cqrs.JSONMarshaler{NewUUID: nil, GenerateName: nil},
		serviceName:   serviceName.String(),
	}
}

var ProvidePublisherFactory = NewPublisherFactory

func (f *PublisherFactory) Create(tx pgx.Tx) (pgrepo.EventPublisher, error) {
	sqlPublisher, err := sql.NewPublisher(
		sql.TxFromPgx(tx),
		sql.PublisherConfig{
			SchemaAdapter:        f.schemaAdapter,
			AutoInitializeSchema: false,
		},
		watermill.NopLogger{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create outbox sql publisher: %w", err)
	}
	publisher := forwarder.NewPublisher(sqlPublisher, forwarder.PublisherConfig{
		ForwarderTopic: ForwarderTopic,
		Marshaler:      nil,
	})
	return &Publisher{
		cfg:       f.cfg,
		publisher: wotel.NewNamedPublisherDecorator(f.serviceName, publisher),
		marshaler: f.marshaler,
	}, nil
}

// Publisher writes events into the outbox table of one transaction; the forwarder relays them to Kafka.
type Publisher struct {
	cfg       *config.DomainEvent
	publisher message.Publisher
	marshaler cqrs.JSONMarshaler
}

var _ pgrepo.EventPublisher = (*Publisher)(nil)

func (p *Publisher) Publish(ctx context.Context, events ...domain.Event) error {
	for _, event := range events {
		topic, err := component.DomainEventToTopic(event)
		if err != nil {
			return err
		}
		msg, err := p.toMessage(ctx, event)
		if err != nil {
			return fmt.Errorf("failed to build outbox message for %T: %w", event, err)
		}
		if err := p.publisher.Publish(topic, msg); err != nil {
			return fmt.Errorf("failed to write %T to outbox: %w", event, err)
		}
	}
	return nil
}

func (p *Publisher) toMessage(ctx context.Context, event domain.Event) (*message.Message, error) {
	payload, partitionKey, err := toPayload(event)
	if err != nil {
		return nil, err
	}
	msg, err := p.marshaler.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg.UUID = event.GetID().String()
	msg.SetContext(ctx)
	msg.Metadata.Set(p.cfg.MessageMetadataAggregateIDKey, event.GetAggregateID())
	msg.Metadata.Set(p.cfg.MessageMetadataUserIDKey, event.GetUserID())
	msg.Metadata.Set(p.cfg.MessageMetadataPartitionKeyKey, partitionKey)
	return msg, nil
}

// toPayload maps the integration events onto their api/share contract; every other domain event
// is published as is. Integration events are keyed by workspace so consumers see them in order.
func toPayload(event domain.Event) (any, string, error) {
	switch e := event.(type) {
	case *domain.SubscriptionStateChangedEvent:
		payload, err := toShareSubscriptionChanged(e)
		return payload, e.WorkspaceID.String(), err
	default:
		return event, event.GetAggregateID(), nil
	}
}
