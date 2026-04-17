package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type NotifyWorkspaceRenamed struct {
	WorkspaceID   uuid.UUID
	UserID        string
	Name          string
	CorrelationID uuid.UUID
}

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

func (h *NotifyWorkspaceRenamedHandler) Handle(ctx context.Context, params *NotifyWorkspaceRenamed) error {
	return h.workspaceEventPublisher.Publish(ctx, params.WorkspaceID, params.UserID, &WorkspaceEventWorkspaceRenamed{
		workspaceEvent[note.WorkspaceRenamedEventEvent]{
			Id:    params.CorrelationID,
			Event: note.WorkspaceRenamedEventEventWorkspaceRenamedEvent,
			Data: note.WorkspaceRenamedEventData{
				Id:   &params.WorkspaceID,
				Name: params.Name,
			},
		},
	})
}
