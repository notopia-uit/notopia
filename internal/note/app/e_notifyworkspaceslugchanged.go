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
	slog.DebugContext(ctx, "Handling workspace slug changed event", slog.String("workspace_id", params.AggregateID.String()))
	err := h.workspaceEventPublisher.Publish(ctx, params.AggregateID, params.UserID, &WorkspaceEventWorkspaceSlugChanged{
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
		slog.ErrorContext(ctx, "failed to publish workspace slug changed event", slog.String("workspace_id", params.AggregateID.String()), slog.Any("error", err))
		return err
	}
	slog.InfoContext(ctx, "Workspace slug changed event published", slog.String("workspace_id", params.AggregateID.String()))
	return nil
}
