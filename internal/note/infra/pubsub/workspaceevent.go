package pubsub

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/redis/go-redis/v9"
)

// TODO: If have time, try https://github.com/stong1994/watermill-rediszset, because we only need pubsub, not stream
func NewWorkspaceEvent(
	logger watermill.LoggerAdapter,
	marshaler cqrs.CommandEventMarshaler,
	redisClient *RedisClient,
) (*pubsub.WorkspaceEvent, error) {
	pubisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:        (*redis.Client)(redisClient),
		DefaultMaxlen: 10000,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis publisher: %w", err)
	}
	subcriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        (*redis.Client)(redisClient),
		ConsumerGroup: "workspace-event-processor",
	}, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis subscriber: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal message router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer)

	eventBus, err := cqrs.NewEventBusWithConfig(pubisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events:workspace" + params.EventName, nil
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
				return "events:workspace" + params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return subcriber, nil
			},
			Marshaler: marshaler,
			Logger:    logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal event processor: %w", err)
	}

	return pubsub.NewWorkspaceEvent(
		eventBus,
		eventProcessor,
		router,
		pubisher,
		subcriber,
	), nil
}

var ProvideWorkspaceEvent = NewWorkspaceEvent
