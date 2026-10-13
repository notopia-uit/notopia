package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type WorkspaceDeletedDomainToIntegrationEventHandler struct {
	integrationPublisher IntegrationPublisher
}

func NewWorkspaceDeletedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
) *WorkspaceDeletedDomainToIntegrationEventHandler {
	return &WorkspaceDeletedDomainToIntegrationEventHandler{integrationPublisher: integrationPublisher}
}

var ProvideWorkspaceDeletedDomainToIntegrationEventHandler = NewWorkspaceDeletedDomainToIntegrationEventHandler

func (h *WorkspaceDeletedDomainToIntegrationEventHandler) Handle(ctx context.Context, event *domain.WorkspaceDeletedEvent) error {
	if err := h.integrationPublisher.Publish(ctx, IntegrationEventWorkspaceDeleted{ID: event.AggregateID}); err != nil {
		return fmt.Errorf("failed to publish workspace deleted integration event: %w", err)
	}
	slog.InfoContext(ctx, "Workspace deleted integration event published", slog.String("workspace_id", event.AggregateID.String()))
	return nil
}
