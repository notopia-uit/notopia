package integrationevent

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/notopia-uit/notopia/internal/note/app"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type IntegrationEvent struct {
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	app            *app.Server
}

func NewIntegrationEvent(
	cfg *commonconfig.Kafka,
	app *app.Server,
	tracer kafka.SaramaTracer,
	logger watermill.LoggerAdapter,
	marshaler *cqrs.JSONMarshaler,
) (*IntegrationEvent, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration event controller router: %w", err)
	}
	router.AddMiddleware(
		middleware.Recoverer,
	)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return "events.integration." + params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return kafka.NewSubscriber(
					kafka.SubscriberConfig{
						Brokers:       cfg.Brokers,
						ConsumerGroup: cfg.ConsumerGroup + "." + params.HandlerName,
						Tracer:        tracer,
					},
					logger,
				)
			},
			Marshaler: marshaler,
			Logger:    logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration event processor: %w", err)
	}

	ie := &IntegrationEvent{
		eventProcessor: eventProcessor,
		router:         router,
		app:            app,
	}
	if err := ie.registerHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register handlers for integration event processor: %w", err)
	}

	return ie, nil
}

var ProvideIntegrationEvent = NewIntegrationEvent

func (c *IntegrationEvent) registerHandlers() error {
	if err := c.eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"DocumentCommittedHandler",
			c.app.IntegrationEvents.DocumentCommittedHandler.Handle,
		),
	); err != nil {
		return fmt.Errorf("failed to add event handlers to integration event processor: %w", err)
	}
	return nil
}

func (c *IntegrationEvent) Run(ctx context.Context) error {
	if err := c.router.Run(ctx); err != nil {
		return fmt.Errorf("integration event controller router stopped with error: %w", err)
	}
	return nil
}

func (c *IntegrationEvent) Close() error {
	if err := c.router.Close(); err != nil {
		return fmt.Errorf("failed to close integration event controller router: %w", err)
	}
	return nil
}
