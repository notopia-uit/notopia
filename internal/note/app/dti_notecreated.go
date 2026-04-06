package app

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type DomainEventToIntegrationEventNoteCreatedHandler struct {
	integrationPublisher IntegrationPublisher
}

var _ DomainEventToIntegrationEventHandler[*domain.NoteCreatedEvent] = (*DomainEventToIntegrationEventNoteCreatedHandler)(nil)

func NewDomainEventToIntegrationEventNoteCreatedHandler(
	integrationPublisher IntegrationPublisher,
) *DomainEventToIntegrationEventNoteCreatedHandler {
	return &DomainEventToIntegrationEventNoteCreatedHandler{
		integrationPublisher: integrationPublisher,
	}
}

func (h *DomainEventToIntegrationEventNoteCreatedHandler) Handle(ctx context.Context, event *domain.NoteCreatedEvent) error {
	integrationEvent := IntegrationEventNoteCreated{
		Id:   &event.AggregateID,
		Icon: event.Icon,
		Name: event.Name,
	}
	if err := h.integrationPublisher.Publish(ctx, integrationEvent); err != nil {
		return fmt.Errorf("failed to publish the converted note created event to the integration publisher: %w", err)
	}
	return nil
}
