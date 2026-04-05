package domainevent

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"golang.org/x/sync/errgroup"
)

type Processor struct {
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	gochannel      *gochannel.GoChannel
	forwarder      *forwarder.Forwarder
}

func NewProcessor(
	logger watermill.LoggerAdapter,
	pgxConn sql.Conn,
	marshaller *cqrs.JSONMarshaler,
) (*Processor, error) {
	channel := gochannel.NewGoChannel(gochannel.Config{}, logger)
	sqlSubcriber, err := sql.NewSubscriber(
		sql.BeginnerFromPgx(pgxConn),
		sql.SubscriberConfig{
			PollInterval:     time.Second,
			InitializeSchema: true,
			SchemaAdapter:    sql.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL subscriber: %w", err)
	}
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}
	forwarder, err := forwarder.NewForwarder(sqlSubcriber, channel, logger, forwarder.Config{
		Router: router,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create forwarder: %w", err)
	}
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return params.EventName, nil
		},
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return channel, nil
		},
		Marshaler: marshaller,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event processor: %w", err)
	}

	return &Processor{
		eventProcessor: eventProcessor,
		router:         router,
		gochannel:      channel,
		forwarder:      forwarder,
	}, nil
}

var ProvideProcessor = NewProcessor

var _ app.DomainEventProcessor = (*Processor)(nil)

func (b *Processor) RegisterHandlers(handlers ...app.DomainEventHandler[domain.Event]) error {
	adaptedHandlers := make([]cqrs.EventHandler, len(handlers))
	for i, handler := range handlers {
		adaptedHandlers[i] = &adapter[domain.Event]{handler: handler}
	}
	if err := b.eventProcessor.AddHandlers(adaptedHandlers...); err != nil {
		return fmt.Errorf("failed to register domain event handlers: %w", err)
	}
	return nil
}

func (b *Processor) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := b.router.Run(ctx); err != nil {
			return fmt.Errorf("failed to run domain event bus: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := b.forwarder.Run(ctx); err != nil {
			return fmt.Errorf("failed to run domain event forwarder: %w", err)
		}
		return nil
	})
	return nil
}

func (b *Processor) Close() error {
	if err := b.router.Close(); err != nil {
		return fmt.Errorf("failed to close domain event bus: %w", err)
	}
	return nil
}
