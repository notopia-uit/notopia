package app

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type NotifyWorkspaceSlugChangedHandler struct {
	workspaceEventPublisher WorkspaceEventPublisher
}

func NewNotifyWorkspaceSlugChangedHandler(
	workspaceEventPublisher WorkspaceEventPublisher,
) *NotifyWorkspaceSlugChangedHandler {
	return &NotifyWorkspaceSlugChangedHandler{
		workspaceEventPublisher: workspaceEventPublisher,
	}
}

var ProvideNotifyWorkspaceSlugChangedHandler = NewNotifyWorkspaceSlugChangedHandler

func (h *NotifyWorkspaceSlugChangedHandler) Handle(ctx context.Context, params *domain.WorkspaceSlugChangedEvent) error {
	slog.DebugContext(ctx, "Handling notify workspace slug changed event", slog.String("workspace_id", params.AggregateID.String()))
	err := h.workspaceEventPublisher.Publish(ctx, &WorkspaceEventPublishParams{
		WorkspaceID: params.AggregateID,
		UserID:      params.UserID,
		SessionID:   "",
	}, &WorkspaceEventWorkspaceSlugChanged{
		workspaceEvent[note.WorkspaceSlugChangedEventEvent]{
			Id:    params.ID,
			Event: note.WorkspaceSlugChangedEventEventWorkspaceSlugChangedEvent,
			Data: note.WorkspaceSlugChangedEventData{
				Id:   &params.AggregateID,
				Slug: params.Slug,
			},
		},
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Notify workspace slug changed event published", slog.String("workspace_id", params.AggregateID.String()))
	return nil
}
