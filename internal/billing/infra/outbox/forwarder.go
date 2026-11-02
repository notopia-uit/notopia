package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/billing/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
	"github.com/notopia-uit/notopia/pkg/otel"
)

// Forwarder relays outbox rows committed by the unit of work to Kafka.
type Forwarder struct {
	fwd *forwarder.Forwarder
}

func NewForwarder(
	pool *pgxpool.Pool,
	kafkaCfg *commonconfig.Kafka,
	domainEventCfg *config.DomainEvent,
	schemaAdapter sql.SchemaAdapter,
	offsetsAdapter sql.OffsetsAdapter,
	tracer kafka.SaramaTracer,
	logger watermill.LoggerAdapter,
	serviceName metadata.ServiceName,
	_ otel.Global,
) (*Forwarder, func(), error) {
	partitionKey := domainEventCfg.MessageMetadataPartitionKeyKey
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: kafkaCfg.Brokers,
			Marshaler: kafka.NewWithPartitioningMarshaler(func(_ string, msg *message.Message) (string, error) {
				return msg.Metadata.Get(partitionKey), nil
			}),
			OverwriteSaramaConfig: nil,
			OTELEnabled:           false,
			Tracer:                tracer,
		},
		logger,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create outbox kafka publisher: %w", err)
	}
	publisher := wotel.NewNamedPublisherDecorator(serviceName.String(), kafkaPublisher)
	cleanup := func() {
		if err := publisher.Close(); err != nil {
			slog.Error("failed to close outbox kafka publisher", slog.Any("error", err))
		}
	}

	subscriber, err := sql.NewSubscriber(
		sql.BeginnerFromPgx(pool),
		sql.SubscriberConfig{
			PollInterval:     time.Second,
			InitializeSchema: false,
			SchemaAdapter:    schemaAdapter,
			OffsetsAdapter:   offsetsAdapter,
		},
		logger,
	)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to create outbox sql subscriber: %w", err)
	}
	router, err := message.NewRouter(message.RouterConfig{CloseTimeout: 0}, logger)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to create outbox router: %w", err)
	}
	router.AddMiddleware(middleware.Recoverer, wotel.Trace())
	fwd, err := forwarder.NewForwarder(subscriber, publisher, logger, forwarder.Config{
		ForwarderTopic:      ForwarderTopic,
		Middlewares:         nil,
		CloseTimeout:        0,
		AckWhenCannotUnwrap: false,
		Router:              router,
		Marshaler:           nil,
	})
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to create outbox forwarder: %w", err)
	}
	return &Forwarder{fwd: fwd}, cleanup, nil
}

var ProvideForwarder = NewForwarder

// Run blocks until ctx is canceled, which closes the forwarder's router.
func (f *Forwarder) Run(ctx context.Context) error {
	if err := f.fwd.Run(ctx); err != nil {
		return fmt.Errorf("failed to run outbox forwarder: %w", err)
	}
	return nil
}
