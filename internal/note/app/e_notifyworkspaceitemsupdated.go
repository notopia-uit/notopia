package app

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

// Because workspace items includes note and folder, so, we share the same debounce
type NotifyWorkspaceItemsUpdatedHandler struct {
	workspaceEventHub       WorkspaceEventHub
	debouncers              sync.Map
	debounceDuration        time.Duration
	noteRepo                domain.NoteRepo
	folderRepo              domain.FolderRepo
	workspaceEventPublisher WorkspaceEventPublisher
}

func NewNotifyWorkspaceItemsUpdatedHandler(
	workspaceEventHub WorkspaceEventHub,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
) *NotifyWorkspaceItemsUpdatedHandler {
	return &NotifyWorkspaceItemsUpdatedHandler{
		workspaceEventHub: workspaceEventHub,
		debouncers:        sync.Map{},
		debounceDuration:  1 * time.Second,
		noteRepo:          noteRepo,
		folderRepo:        folderRepo,
	}
}

var ProvideNotifyWorkspaceItemsUpdatedHandler = NewNotifyWorkspaceItemsUpdatedHandler

// register this to the topics of domain folder
func (h *NotifyWorkspaceItemsUpdatedHandler) HandleFolder(
	ctx context.Context,
	workspaceID uuid.UUID,
	folderID uuid.UUID,
) error {
	return nil
}

// register this to the topics of domain note
func (h *NotifyWorkspaceItemsUpdatedHandler) HandleNote(
	ctx context.Context,
	workspaceID uuid.UUID,
	noteID uuid.UUID,
) error {
	return nil
}
