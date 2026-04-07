package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type NotifyWorkspaceItemsUpdatedFolder struct {
	WorkspaceID uuid.UUID
	UserID      string
	FolderID    uuid.UUID
}

type NotifyWorkspaceItemsUpdatedNote struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteID      uuid.UUID
}

// Because workspace items includes note and folder, so, we share the same debounce
type NotifyWorkspaceItemsUpdatedHandler struct {
	debouncers              sync.Map
	debounceDuration        time.Duration
	noteRepo                domain.NoteRepo
	folderRepo              domain.FolderRepo
	workspaceEventPublisher WorkspaceEventPublisher
}

func NewNotifyWorkspaceItemsUpdatedHandler(
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	workspaceEventPublisher WorkspaceEventPublisher,
) *NotifyWorkspaceItemsUpdatedHandler {
	return &NotifyWorkspaceItemsUpdatedHandler{
		debouncers:              sync.Map{},
		debounceDuration:        1 * time.Second,
		noteRepo:                noteRepo,
		folderRepo:              folderRepo,
		workspaceEventPublisher: workspaceEventPublisher,
	}
}

var ProvideNotifyWorkspaceItemsUpdatedHandler = NewNotifyWorkspaceItemsUpdatedHandler

// register this to the topics of domain folder
func (h *NotifyWorkspaceItemsUpdatedHandler) HandleFolder(
	ctx context.Context,
	params *NotifyWorkspaceItemsUpdatedFolder,
) {
	return
}

// register this to the topics of domain note
func (h *NotifyWorkspaceItemsUpdatedHandler) HandleNote(
	ctx context.Context,
	params *NotifyWorkspaceItemsUpdatedNote,
) {
	return
}

func (h *NotifyWorkspaceItemsUpdatedHandler) executeDebounced(workspaceID uuid.UUID, userID string) {
	val, _ := h.debouncers.LoadOrStore(workspaceID, debounce.New(h.debounceDuration))
	debouncer := val.(func(func()))

	debouncer(func() {
		publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := h.publishWorkspaceUpdate(publishCtx, workspaceID, userID)
		if err != nil {
			slog.Error(
				"failed to publish workspace update",
				slog.String("workspaceID", workspaceID.String()),
				slog.Any("error", err),
			)
		}

		h.debouncers.Delete(workspaceID)
	})
}

func (h *NotifyWorkspaceItemsUpdatedHandler) publishWorkspaceUpdate(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) error {
	id, err := uuid.NewV7()
	if err != nil {
		return errors.New("failed to generate event ID")
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
	h.workspaceEventPublisher.Publish(
		ctx,
		workspaceID,
		userID,
		event,
	)
	return nil
}
