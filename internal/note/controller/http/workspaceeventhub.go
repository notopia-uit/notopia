package http

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
)

type (
	WorkspaceID = uuid.UUID
	UserID      = string
)

type WorkspaceEventHub struct {
	clients map[WorkspaceID]map[UserID]chan<- []byte
	mu      sync.RWMutex
	pubsub  *pubsub.WorkspaceEvent
}

func NewWorkspaceEventHub(pubsub *pubsub.WorkspaceEvent) *WorkspaceEventHub {
	return &WorkspaceEventHub{
		clients: make(map[WorkspaceID]map[UserID]chan<- []byte),
		pubsub:  pubsub,
	}
}

var ProvideWorkspaceEventHub = NewWorkspaceEventHub

func (h *WorkspaceEventHub) Subscribe(workspaceID WorkspaceID, userID UserID, ch chan<- []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[workspaceID]; !exists {
		h.clients[workspaceID] = make(map[UserID]chan<- []byte)
	}
	h.clients[workspaceID][userID] = ch
}

func (h *WorkspaceEventHub) Unsubscribe(workspaceID WorkspaceID, userID UserID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if users, exists := h.clients[workspaceID]; exists {
		delete(users, userID)
		if len(users) == 0 {
			delete(h.clients, workspaceID)
		}
	}
}

func (h *WorkspaceEventHub) Run(ctx context.Context) error {
	messages, err := h.pubsub.Subcribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to workspace events: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-messages:
			if !ok {
				slog.InfoContext(ctx, "workspace event channel closed")
				return nil
			}
			h.handleMessage(ctx, msg)
		}
	}
}

func (h *WorkspaceEventHub) handleMessage(ctx context.Context, msg *message.Message) {
	defer msg.Ack()

	workspaceIDString := msg.Metadata.Get(pubsub.MetadataWorkspaceIDKey)
	workspaceID, err := uuid.Parse(workspaceIDString)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"invalid workspace ID in message metadata",
			slog.String("workspace_id", workspaceIDString),
			slog.String("error", err.Error()),
		)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if users, exists := h.clients[workspaceID]; exists {
		for _, ch := range users {
			select {
			case ch <- msg.Payload:
			case <-ctx.Done():
				return
			}
		}
	}
}
