package integrationpublisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

type Publisher message.Publisher

type IntegrationPublisher struct {
	publisher Publisher
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
		transformedEvent, err := transformIntegrationEvent(event)
		if err != nil {
			return fmt.Errorf("cannot convert event to integration event: %w", err)
		}
		topic, err := component.IntegrationEventToTopic(event)
		if err != nil {
			return fmt.Errorf("cannot get topic for integration event: %w", err)
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

func transformIntegrationEvent(event app.IntegrationEvent) (any, error) {
	switch e := event.(type) {
	case app.IntegrationEventNoteCreated:
		var icon *string
		if e.Icon != "" {
			icon = &e.Icon
		}
		return &share.NoteCreatedEvent{
			Id:          e.ID,
			WorkspaceId: e.WorkspaceID,
			Icon:        icon,
			Name:        e.Name,
		}, nil
	case app.IntegrationEventNoteDeleted:
		return &share.NoteDeletedEvent{
			Id: e.ID,
		}, nil
	case app.IntegrationEventNoteUpdated:
		var icon *string
		if e.Icon != "" {
			icon = &e.Icon
		}
		trashed, err := toTrashed(&e.Trashed)
		if err != nil {
			return nil, fmt.Errorf("failed to convert trashed: %w", err)
		}
		return &share.NoteUpdatedEvent{
			Id:          e.ID,
			WorkspaceId: e.WorkspaceID,
			Name:        e.Name,
			Icon:        icon,
			FolderId:    e.FolderID,
			FolderName:  e.FolderName,
			Trashed:     trashed,
			UpdatedAt:   e.UpdatedAt,
		}, nil
	}
	return nil, fmt.Errorf("unknown integration event type: %T", event)
}

func toTrashed(trashed *app.Trashed) (*share.Trashed, error) {
	if trashed == nil || !trashed.IsTrashed() {
		return nil, nil
	}
	trashedBy, err := toTrashedBy(trashed.By)
	if err != nil {
		return nil, fmt.Errorf("failed to convert trashed by: %w", err)
	}
	return &share.Trashed{
		At: trashed.At,
		By: trashedBy,
	}, nil
}

func toTrashedBy(trashedBy app.TrashedBy) (share.TrashedBy, error) {
	switch trashedBy {
	case app.TrashedByPurpose:
		return share.Parent, nil
	case app.TrashedByParent:
		return share.Parent, nil
	case app.TrashedByUnspecified:
		return "", fmt.Errorf("invalid trashed by value: %v", trashedBy)
	}
	return "", fmt.Errorf("unknown trashed by value: %v", trashedBy)
}
