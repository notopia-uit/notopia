package pubsub

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/notopia-uit/notopia/internal/note/app"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type IntegrationPublisher struct {
	bus *cqrs.EventBus
}

var _ app.IntegrationPublisher = (*IntegrationPublisher)(nil)

func NewIntegrationPublisher(
	cfg commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	tracer kafka.SaramaTracer,
	marshaller *cqrs.JSONMarshaler,
) (*IntegrationPublisher, error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}
	bus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			event, ok := params.Event.(app.IntegrationEvent)
			if !ok {
				return "", fmt.Errorf("not integration event: %T", params.Event)
			}
			return "events.integration." + event.Type().String(), nil
		},
		Marshaler: marshaller,
		Logger:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event bus: %w", err)
	}
	return &IntegrationPublisher{
		bus: bus,
	}, nil
}

var ProvideIntegrationPublisher = NewIntegrationPublisher

func (p *IntegrationPublisher) Publish(ctx context.Context, events ...app.IntegrationEvent) error {
	for _, event := range events {
		if err := p.bus.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish integration event: %w", err)
		}
	}
	return nil
}
