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
	workspaceEventHub WorkspaceEventHub
	debouncers        sync.Map // workspaceID -> func(func())
	pendingEvents     sync.Map // workspaceID -> latest event
	debounceDuration  time.Duration
	noteRepo          domain.NoteRepo
	folderRepo        domain.FolderRepo
}

func NewNotifyWorkspaceItemsUpdatedHandler(
	workspaceEventHub WorkspaceEventHub,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
) *NotifyWorkspaceItemsUpdatedHandler {
	return &NotifyWorkspaceItemsUpdatedHandler{
		workspaceEventHub: workspaceEventHub,
		debouncers:        sync.Map{},
		pendingEvents:     sync.Map{},
		debounceDuration:  1 * time.Second,
		noteRepo:          noteRepo,
		folderRepo:        folderRepo,
	}
}

var ProvideNotifyWorkspaceItemsUpdatedHandler = NewNotifyWorkspaceItemsUpdatedHandler

// register this to the topics of domain folder
func (h *NotifyWorkspaceItemsUpdatedHandler) HandleFolder(
	ctx context.Context,
	folderID uuid.UUID,
) error {
	return nil
}

// register this to the topics of domain note
func (h *NotifyWorkspaceItemsUpdatedHandler) HandleNote(
	ctx context.Context,
	noteID uuid.UUID,
) error {
	return nil
}
