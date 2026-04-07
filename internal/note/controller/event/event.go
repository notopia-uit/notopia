package event

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

// This include integration (from share package) and domain event, setup for event processor
// If not use with event processor but via subscriber directly, not need to declare this
func eventToTopic(event any) (string, bool) {
	switch e := event.(type) {
	case domain.Event:
		return component.DomainEventToTopic(e)
	case *share.DocumentCommittedEvent:
		return "events.integration.document.document.committed", true
	}
	return "", false
}

type Event struct {
	subcriber      message.Subscriber
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	app            *app.Server
}

func NewEvent(
	cfg *commonconfig.Kafka,
	app *app.Server,
	tracer kafka.SaramaTracer,
	logger watermill.LoggerAdapter,
	marshaler *cqrs.JSONMarshaler,
) (*Event, error) {
	subcriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       cfg.Brokers,
			ConsumerGroup: cfg.ConsumerGroup,
			Tracer:        tracer,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller subscriber: %w", err)
	}
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller router: %w", err)
	}
	retryMiddleware := middleware.Retry{
		MaxRetries:      3,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2,
		Logger:          logger,
	}
	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Recoverer,
		retryMiddleware.Middleware,
	)
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				if topic, ok := eventToTopic(params.EventHandler.NewEvent()); ok { // Watermill doesn't pass the new event for me??, so I create a new event again
					return topic, nil
				}
				return "", fmt.Errorf("unknown event name %s", params.EventName)
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
		return nil, fmt.Errorf("failed to create event controller event processor: %w", err)
	}
	event := &Event{
		subcriber:      subcriber,
		eventProcessor: eventProcessor,
		router:         router,
		app:            app,
	}
	if err := event.setup(); err != nil {
		return nil, fmt.Errorf("failed to setup event controller: %w", err)
	}
	return event, nil
}

var ProvideEvent = NewEvent

func (e *Event) setup() error {
	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"DocumentCommittedHandler",
		e.documentCommittedHandler,
	)); err != nil {
		return fmt.Errorf("failed to add event handler: %w", err)
	}
	// TODO: because watermill doesn't support kafka regex (IBM/sarama)
	// So, we will need to for loop all topic we have, (for note, and folder)
	// And handle for workspace item updated
	return nil
}

func (e *Event) Run(ctx context.Context) error {
	if err := e.router.Run(ctx); err != nil {
		return fmt.Errorf("failed to run event controller router: %w", err)
	}
	return nil
}

func (e *Event) IsRunning() bool {
	return e.router.IsRunning()
}
