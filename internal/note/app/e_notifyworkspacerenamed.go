package app

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type NotifyWorkspaceRenamedHandler struct {
	workspaceEventPublisher WorkspaceEventPublisher
}

func NewNotifyWorkspaceRenamedHandler(
	workspaceEventPublisher WorkspaceEventPublisher,
) *NotifyWorkspaceRenamedHandler {
	return &NotifyWorkspaceRenamedHandler{
		workspaceEventPublisher: workspaceEventPublisher,
	}
}

var ProvideNotifyWorkspaceRenamedHandler = NewNotifyWorkspaceRenamedHandler

func (h *NotifyWorkspaceRenamedHandler) Handle(ctx context.Context, params *domain.WorkspaceRenamedEvent) error {
	slog.DebugContext(ctx, "Handling workspace renamed event", slog.String("workspace_id", params.AggregateID.String()))
	err := h.workspaceEventPublisher.Publish(ctx, params.AggregateID, params.UserID, &WorkspaceEventWorkspaceRenamed{
		workspaceEvent[note.WorkspaceRenamedEventEvent]{
			Id:    params.ID,
			Event: note.WorkspaceRenamedEventEventWorkspaceRenamedEvent,
			Data: note.WorkspaceRenamedEventData{
				Id:   &params.AggregateID,
				Name: params.Name,
			},
		},
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to publish workspace renamed event", slog.String("workspace_id", params.AggregateID.String()), slog.Any("error", err))
		return err
	}
	slog.InfoContext(ctx, "Workspace renamed event published", slog.String("workspace_id", params.AggregateID.String()))
	return nil
}
