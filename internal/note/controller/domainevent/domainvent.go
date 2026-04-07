package domainevent

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/notopia-uit/notopia/internal/note/app"
)

type Subcriber message.Subscriber

// Actually it is the processor (receive and handling)
type DomainEvent struct {
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	subcriber      Subcriber
}

func NewDomainEvent(
	logger watermill.LoggerAdapter,
	subscriber Subcriber,
	marshaller *cqrs.JSONMarshaler,
	app *app.Server,
) (*DomainEvent, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain event controller router: %w", err)
	}
	router.AddMiddleware(middleware.Recoverer)
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return params.EventName, nil
		},
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return subscriber, nil
		},
		Marshaler: marshaller,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create domain event controller event processor: %w", err)
	}

	if err := eventProcessor.AddHandlers(
	// cqrs.NewEventHandler(
	// 			"NoteCreatedDomainToIntegration",
	// ),
	); err != nil {
		return nil, fmt.Errorf("failed to add domain event controller handlers to event processor: %w", err)
	}

	return &DomainEvent{
		eventProcessor: eventProcessor,
		router:         router,
		subcriber:      subscriber,
	}, nil
}

var ProvideDomainEvent = NewDomainEvent

func (b *DomainEvent) Run(ctx context.Context) error {
	if err := b.router.Run(ctx); err != nil {
		return fmt.Errorf("failed to run domain event controller router: %w", err)
	}
	return nil
}

func (b *DomainEvent) Close() error {
	if err := b.router.Close(); err != nil {
		return fmt.Errorf("failed to close domain event controller router: %w", err)
	}
	return nil
}
