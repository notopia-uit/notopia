package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type NoteDeletedDomainToIntegrationEventHandler struct {
	integrationPublisher IntegrationPublisher
}

func NewNoteDeletedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
) *NoteDeletedDomainToIntegrationEventHandler {
	return &NoteDeletedDomainToIntegrationEventHandler{integrationPublisher: integrationPublisher}
}

var ProvideNoteDeletedDomainToIntegrationEventHandler = NewNoteDeletedDomainToIntegrationEventHandler

func (h *NoteDeletedDomainToIntegrationEventHandler) Handle(ctx context.Context, event *domain.NoteDeletedEvent) error {
	if err := h.integrationPublisher.Publish(ctx, IntegrationEventNoteDeleted{ID: event.AggregateID}); err != nil {
		return fmt.Errorf("failed to publish note deleted integration event: %w", err)
	}
	slog.InfoContext(ctx, "Note deleted integration event published", slog.String("note_id", event.AggregateID.String()))
	return nil
}
