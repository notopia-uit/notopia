package app

import (
	"context"

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
	return h.workspaceEventPublisher.Publish(ctx, params.AggregateID, params.UserID, &WorkspaceEventWorkspaceSlugChanged{
		workspaceEvent[note.WorkspaceSlugChangedEventEvent]{
			Id:    params.ID,
			Event: note.WorkspaceSlugChangedEventEventWorkspaceSlugChangedEvent,
			Data: note.WorkspaceSlugChangedEventData{
				Id:   &params.AggregateID,
				Slug: params.Slug,
			},
		},
	})
}
