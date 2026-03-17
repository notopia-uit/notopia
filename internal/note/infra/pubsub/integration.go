package pubsub

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type Integration struct {
	eventBus       *cqrs.EventBus
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	publisher      message.Publisher
}

var _ pubsub.Integration = (*Integration)(nil)

func NewIntegration(
	cfg *commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	marshaler cqrs.CommandEventMarshaler,
) (*Integration, error) {
	tracer := kafka.NewOTELSaramaTracer()

	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration event publisher: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration event router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer)

	eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Marshaler: marshaler,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create integration event bus: %w", err)
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return "events." + params.EventName, nil
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

	return &Integration{
		eventBus:       eventBus,
		eventProcessor: eventProcessor,
		router:         router,
		publisher:      publisher,
	}, nil
}

var ProvideIntegration = NewIntegration
