package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

// NOTE: Struggling with generic, this package should have the constructor for each
// Struggling with external struct also, maybe this should own the struct definition

type WorkspaceEventPublisher interface {
	Publish(
		ctx context.Context,
		workspaceID uuid.UUID,
		userID string,
		events ...WorkspaceEvent,
	) error
}

type WorkspaceEventSubscriber interface {
	Subscribe(
		ctx context.Context,
		workspaceID uuid.UUID,
		userID string,
	) (<-chan WorkspaceEvent, error)
}

// Or, we just marshall the event to a string json string, instead marshall and unmarshall
type WorkspaceEventHub interface {
	WorkspaceEventPublisher
	WorkspaceEventSubscriber
}

type WorkspaceEvent interface {
	isWorkspaceEvent()
	GetID() uuid.UUID
	GetEvent() string
}

type workspaceEvent[E ~string] struct {
	Id    uuid.UUID `json:"id"`
	Event E         `json:"event"`
	Data  any       `json:"data"`
}

var _ WorkspaceEvent = (*workspaceEvent[string])(nil)

func (e workspaceEvent[E]) isWorkspaceEvent() {}
func (e workspaceEvent[E]) GetID() uuid.UUID  { return e.Id }
func (e workspaceEvent[E]) GetEvent() string  { return string(e.Event) }

type WorkspaceEventWorkspaceItemsUpdated struct {
	workspaceEvent[note.WorkspaceItemsUpdatedEventEvent]
}

type WorkspaceEventMembersUpdated struct {
	workspaceEvent[note.WorkspaceMembersUpdatedEventEvent]
}

type WorkspaceEventWorkspaceUpdated struct {
	workspaceEvent[note.WorkspaceUpdatedEventEvent]
}

type WorkspaceEventWorkspaceDeleted struct {
	workspaceEvent[note.WorkspaceDeletedEventEvent]
}

func NewEmptyWorkspaceEventFromType(t string) (WorkspaceEvent, bool) {
	switch t {
	case string(note.WorkspaceMembersUpdatedEventEventWorkspaceMembersUpdatedEvent):
		//exhaustruct:ignore
		return &WorkspaceEventMembersUpdated{}, true
	case string(note.WorkspaceItemsUpdatedEventEventWorkspaceItemsUpdatedEvent):
		//exhaustruct:ignore
		return &WorkspaceEventWorkspaceItemsUpdated{}, true
	case string(note.WorkspaceUpdatedEventEventWorkspaceUpdatedEvent):
		//exhaustruct:ignore
		return &WorkspaceEventWorkspaceUpdated{}, true
	case string(note.WorkspaceDeletedEventEventWorkspaceDeletedEvent):
		//exhaustruct:ignore
		return &WorkspaceEventWorkspaceDeleted{}, true
	default:
		return nil, false
	}
}
