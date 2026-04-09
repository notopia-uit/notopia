package integrationpublisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func transformIntegrationEvent(event app.IntegrationEvent) (any, bool) {
	switch e := event.(type) {
	case app.IntegrationEventNoteCreated:
		var icon *string
		if e.Icon != "" {
			icon = &e.Icon
		}
		return &share.NoteCreatedEvent{
			Id:   e.ID,
			Icon: icon,
			Name: e.Name,
		}, true
	case app.IntegrationEventNoteDeleted:
		return &share.NoteDeletedEvent{
			Id: e.ID,
		}, true
	case app.IntegrationEventNoteUpdated:
		var icon *string
		if e.Icon != "" {
			icon = &e.Icon
		}
		return &share.NoteUpdatedEvent{
			Id:        e.ID,
			Name:      e.Name,
			Icon:      icon,
			Tags:      e.Tags,
			FolderId:  e.FolderID,
			UpdatedAt: e.UpdatedAt,
		}, true
	}
	return nil, false
}

func getIntegrationEventTopic(event app.IntegrationEvent) (string, bool) {
	switch event.(type) {
	case app.IntegrationEventNoteCreated:
		return "events.integration.note.note.created", true
	case app.IntegrationEventNoteDeleted:
		return "events.integration.note.note.deleted", true
	case app.IntegrationEventNoteUpdated:
		return "events.integration.note.note.updated", true
	}
	return "", false
}

type Publisher message.Publisher

type IntegrationPublisher struct {
	publisher message.Publisher
}

var _ app.IntegrationPublisher = (*IntegrationPublisher)(nil)

func NewIntegrationPublisher(
	publisher Publisher,
) (*IntegrationPublisher, error) {
	return &IntegrationPublisher{
		publisher: publisher,
	}, nil
}

var ProvideIntegrationPublisher = NewIntegrationPublisher

func (p *IntegrationPublisher) Publish(ctx context.Context, events ...app.IntegrationEvent) error {
	for _, event := range events {
		transformedEvent, ok := transformIntegrationEvent(event)
		if !ok {
			return fmt.Errorf("cannot convert event to integration event: %T", event)
		}
		topic, ok := getIntegrationEventTopic(event)
		if !ok {
			return fmt.Errorf("cannot get topic for integration event: %T", event)
		}
		payload, err := json.Marshal(transformedEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal integration event: %w", err)
		}
		msg := message.NewMessageWithContext(ctx, watermill.NewUUID(), payload)
		if err := p.publisher.Publish(topic, msg); err != nil {
			return fmt.Errorf("failed to publish integration event: %w", err)
		}
	}
	return nil
}
