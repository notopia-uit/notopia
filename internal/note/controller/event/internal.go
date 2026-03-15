package event

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/notopia-uit/notopia/internal/note/app"
)

type Internal struct {
	EventBus       *cqrs.EventBus
	EventProcessor *cqrs.EventProcessor
	Router         *message.Router
	App            *app.App
}

// TODO: If scaling, need to use Redis to share between server sent event
func NewInternal(
	logger watermill.LoggerAdapter,
	marshaler cqrs.CommandEventMarshaler,
) (*Internal, error) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal message router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer)

	eventBus, err := cqrs.NewEventBusWithConfig(pubSub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Marshaler: marshaler,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create internal event bus: %w", err)
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return "events." + params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return pubSub, nil
			},
			Marshaler: marshaler,
			Logger:    logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal event processor: %w", err)
	}

	return &Internal{
		EventBus:       eventBus,
		EventProcessor: eventProcessor,
		Router:         router,
	}, nil
}

var ProvideInternal = NewInternal
