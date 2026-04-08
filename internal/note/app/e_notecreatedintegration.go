package app

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type NoteCreatedDomainToIntegrationEventHandler struct {
	integrationPublisher IntegrationPublisher
}

func NewNoteCreatedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
) *NoteCreatedDomainToIntegrationEventHandler {
	return &NoteCreatedDomainToIntegrationEventHandler{
		integrationPublisher: integrationPublisher,
	}
}

func (h *NoteCreatedDomainToIntegrationEventHandler) Handle(ctx context.Context, event *domain.NoteCreatedEvent) error {
	integrationEvent := IntegrationEventNoteCreated{
		ID:   event.AggregateID,
		Icon: event.Icon,
		Name: event.Name,
	}
	if err := h.integrationPublisher.Publish(ctx, integrationEvent); err != nil {
		return fmt.Errorf("failed to publish the converted note created event to the integration publisher: %w", err)
	}
	return nil
}
