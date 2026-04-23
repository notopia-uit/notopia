package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type NotifyWorkspaceItemsUpdated struct {
	UserID      string
	WorkspaceID uuid.UUID
}

// Because workspace items includes note and folder, so, we share the same debounce
type NotifyWorkspaceItemsUpdatedHandler struct {
	debouncers              sync.Map
	debounceDuration        time.Duration
	workspaceEventPublisher WorkspaceEventPublisher
}

func NewNotifyWorkspaceItemsUpdatedHandler(
	workspaceEventPublisher WorkspaceEventPublisher,
) *NotifyWorkspaceItemsUpdatedHandler {
	return &NotifyWorkspaceItemsUpdatedHandler{
		debouncers:              sync.Map{},
		debounceDuration:        1 * time.Second,
		workspaceEventPublisher: workspaceEventPublisher,
	}
}

var ProvideNotifyWorkspaceItemsUpdatedHandler = NewNotifyWorkspaceItemsUpdatedHandler

func (h *NotifyWorkspaceItemsUpdatedHandler) Handle(params *NotifyWorkspaceItemsUpdated) error {
	val, _ := h.debouncers.LoadOrStore(params.WorkspaceID, debounce.New(h.debounceDuration))
	debouncer, ok := val.(func(func()))
	if !ok {
		return fmt.Errorf("failed to assert debouncer for workspaceID: %s", params.WorkspaceID.String())
	}

	debouncer(func() {
		publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := h.publishWorkspaceUpdate(publishCtx, params.WorkspaceID, params.UserID)
		if err != nil {
			slog.Error(
				"notify workspace items updated failed to publish event",
				slog.String("workspaceID", params.WorkspaceID.String()),
				slog.Any("error", err),
			)
		}

		h.debouncers.Delete(params.WorkspaceID)
	})
	return nil
}

func (h *NotifyWorkspaceItemsUpdatedHandler) publishWorkspaceUpdate(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("notify workspace item updated failed to generate event ID: %w", err)
	}
	event := &WorkspaceEventWorkspaceItemsUpdated{
		workspaceEvent[note.WorkspaceItemsUpdatedEventEvent]{
			Id:    id,
			Event: note.WorkspaceItemsUpdatedEventEventWorkspaceItemsUpdatedEvent,
			Data: note.WorkspaceItemsUpdatedEventData{
				WorkspaceId: &workspaceID,
			},
		},
	}
	if err := h.workspaceEventPublisher.Publish(ctx, workspaceID, userID, event); err != nil {
		return fmt.Errorf("notify workspace item updated failed to publish event: %w", err)
	}
	return nil
}
