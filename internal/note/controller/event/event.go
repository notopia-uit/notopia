package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type Event struct {
	EventBus       *cqrs.EventBus
	eventProcessor *cqrs.EventProcessor
	Router         *message.Router
}

func New(
	ctx context.Context,
	consumerGroup string,
	logger *slog.Logger,
	cfg *commonconfig.Kafka,
) (*Event, error) {
	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}
	tracer := kafka.NewOTELSaramaTracer()

	// saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	// saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	watermillLogger := watermill.NewSlogLogger(logger)
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		watermillLogger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka publisher: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create message router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID)
	router.AddMiddleware(middleware.Recoverer)

	eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Marshaler: marshaler,
		Logger:    watermillLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event bus: %w", err)
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
					watermillLogger,
				)
			},
			Marshaler: marshaler,
			Logger:    watermillLogger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event processor: %w", err)
	}

	return &Event{
		EventBus:       eventBus,
		eventProcessor: eventProcessor,
		Router:         router,
	}, nil
}
