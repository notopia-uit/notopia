package pubsub

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/notopia-uit/notopia/internal/note/app"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

func NewIntegrationPubSub(
	cfg *commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	publisher *KafkaPublisher,
	tracer kafka.SaramaTracer,
	marshaler cqrs.CommandEventMarshaler,
) (*app.IntegrationPubSub, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer)

	eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events.integration." + params.EventName, nil
		},
		Marshaler: marshaler,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event bus: %w", err)
	}

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
		return nil, fmt.Errorf("failed to create event processor: %w", err)
	}

	return app.NewIntegrationPubSub(
		eventBus,
		eventProcessor,
		router,
	), nil
}

var ProvideIntegrationPubSub = NewIntegrationPubSub
