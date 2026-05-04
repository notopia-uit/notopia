package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type NoteCreatedDomainToIntegrationEventHandler struct {
	integrationPublisher        IntegrationPublisher
	getWorkspaceByNoteReadModel GetWorkspaceByNoteReadModel
}

func NewNoteCreatedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
	getWorkspaceByNoteReadModel GetWorkspaceByNoteReadModel,
) *NoteCreatedDomainToIntegrationEventHandler {
	return &NoteCreatedDomainToIntegrationEventHandler{
		integrationPublisher:        integrationPublisher,
		getWorkspaceByNoteReadModel: getWorkspaceByNoteReadModel,
	}
}

var ProvideNoteCreatedDomainToIntegrationEventHandler = NewNoteCreatedDomainToIntegrationEventHandler

func (h *NoteCreatedDomainToIntegrationEventHandler) Handle(ctx context.Context, event *domain.NoteCreatedEvent) error {
	slog.DebugContext(ctx, "Handling note created integration event", slog.String("note_id", event.AggregateID.String()))
	workspace, err := h.getWorkspaceByNoteReadModel.GetWorkspaceByNoteID(ctx, event.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get workspace for note: %w", err)
	}
	integrationEvent := IntegrationEventNoteCreated{
		ID:          event.AggregateID,
		WorkspaceID: workspace.ID,
		Icon:        event.Icon,
		Name:        event.Name,
	}
	if err := h.integrationPublisher.Publish(ctx, integrationEvent); err != nil {
		return fmt.Errorf("failed to publish the converted note created event to the integration publisher: %w", err)
	}
	slog.InfoContext(ctx, "Note created integration event published", slog.String("note_id", event.AggregateID.String()))
	return nil
}
