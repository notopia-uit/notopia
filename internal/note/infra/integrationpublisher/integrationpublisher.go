package integrationpublisher

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

func toIntegrationEvent(event app.IntegrationEvent) (any, bool) {
	switch e := event.(type) {
	case app.IntegrationEventNoteCreated:
		return &share.NoteCreatedEvent{
			Id:   &e.ID,
			Icon: e.Icon,
			Name: e.Name,
		}, true
	case app.IntegrationEventNoteDeleted:
		return &share.NoteDeletedEvent{
			Id: &e.ID,
		}, true
	case app.IntegrationEventNoteUpdated:
		var trashed *share.NoteTrashed
		if e.TrashedBy != nil && e.TrashedAt != nil {
			// TODO: This should carefully cast to the right type
			trashed = &share.NoteTrashed{
				TrashedBy: share.TrashedBy(*e.TrashedBy),
				TrashedAt: *e.TrashedAt,
			}
		}
		return &share.NoteUpdatedEvent{
			Id:        &e.ID,
			Name:      e.Name,
			Icon:      e.Icon,
			Tags:      &e.Tags,
			FolderId:  &e.FolderID,
			Trashed:   trashed,
			UpdatedAt: &e.UpdatedAt,
		}, true
	}
	return nil, false
}

type IntegrationPublisher struct {
	bus *cqrs.EventBus
}

var _ app.IntegrationPublisher = (*IntegrationPublisher)(nil)

func NewIntegrationPublisher(
	cfg *commonconfig.Kafka,
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
			return "events.integration." + params.EventName, nil
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
		integrationEvent, ok := toIntegrationEvent(event)
		if !ok {
			return fmt.Errorf("unsupported integration event type: %T", event)
		}
		if err := p.bus.Publish(ctx, integrationEvent); err != nil {
			return fmt.Errorf("failed to publish integration event: %w", err)
		}
	}
	return nil
}
