package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

var (
	EventTypeUnspecified EventType = "UnspecifiedEvent"

	EventTypeFolderCreated            EventType = "FolderCreatedEvent"
	EventTypeFolderDeleted            EventType = "FolderDeletedEvent"
	EventTypeFolderUpdated            EventType = "FolderUpdatedEvent"
	EventTypeFolderMoved              EventType = "FolderMovedEvent"
	EventTypeFolderTrashed            EventType = "FolderTrashedEvent"
	EventTypeFolderRestored           EventType = "FolderRestoredEvent"
	EventTypeFolderPermanentlyDeleted EventType = "FolderPermanentlyDeletedEvent"

	EventTypeNoteCreated            EventType = "NoteCreatedEvent"
	EventTypeNoteDeleted            EventType = "NoteDeletedEvent"
	EventTypeNoteUpdated            EventType = "NoteUpdatedEvent"
	EventTypeNoteMoved              EventType = "NoteMovedEvent"
	EventTypeNoteTrashed            EventType = "NoteTrashedEvent"
	EventTypeNoteRestored           EventType = "NoteRestoredEvent"
	EventTypeNotePermanentlyDeleted EventType = "NotePermanentlyDeletedEvent"

	EventTypeWorkspaceUpdated EventType = "WorkspaceUpdatedEvent"
	EventTypeWorkspaceDeleted EventType = "WorkspaceDeletedEvent"
)

func (t EventType) String() string {
	return string(t)
}

type Event interface {
	GetID() uuid.UUID
	GetOccurredAt() time.Time
	GetAggregateID() uuid.UUID
	GetUserID() string
}

type BaseEvent struct {
	ID          uuid.UUID `json:"id"`
	OccurredAt  time.Time `json:"occurredAt"`
	AggregateID uuid.UUID `json:"aggregateId"`
	UserID      string    `json:"userId"`
}

var _ Event = (*BaseEvent)(nil)

func (e *BaseEvent) GetID() uuid.UUID { return e.ID }

func (e *BaseEvent) GetOccurredAt() time.Time { return e.OccurredAt }

func (e *BaseEvent) GetAggregateID() uuid.UUID { return e.AggregateID }

func (e *BaseEvent) GetUserID() string { return e.UserID }

func NewBaseEvent(aggregateID uuid.UUID, userID string) BaseEvent {
	return BaseEvent{
		ID:          uuid.New(),
		OccurredAt:  time.Now(),
		AggregateID: aggregateID,
		UserID:      userID,
	}
}

func GetEventType(e Event) EventType {
	if e == nil {
		return EventTypeUnspecified
	}
	switch e.(type) {
	case *FolderCreatedEvent:
		return EventTypeFolderCreated
	case *FolderDeletedEvent:
		return EventTypeFolderDeleted
	case *FolderUpdatedEvent:
		return EventTypeFolderUpdated
	case *FolderMovedEvent:
		return EventTypeFolderMoved
	case *FolderTrashedEvent:
		return EventTypeFolderTrashed
	case *FolderRestoredEvent:
		return EventTypeFolderRestored
	case *FolderPermanentlyDeletedEvent:
		return EventTypeFolderPermanentlyDeleted
	case *NoteCreatedEvent:
		return EventTypeNoteCreated
	case *NoteDeletedEvent:
		return EventTypeNoteDeleted
	case *NoteUpdatedEvent:
		return EventTypeNoteUpdated
	case *NoteMovedEvent:
		return EventTypeNoteMoved
	case *NoteTrashedEvent:
		return EventTypeNoteTrashed
	case *NoteRestoredEvent:
		return EventTypeNoteRestored
	case *NotePermanentlyDeletedEvent:
		return EventTypeNotePermanentlyDeleted
	case *WorkspaceUpdatedEvent:
		return EventTypeWorkspaceUpdated
	case *WorkspaceDeletedEvent:
		return EventTypeWorkspaceDeleted
	default:
		return EventTypeUnspecified
	}
}

type FolderCreatedEvent struct {
	BaseEvent
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type FolderDeletedEvent struct {
	BaseEvent
}

type FolderUpdatedEvent struct {
	BaseEvent
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type NoteCreatedEvent struct {
	BaseEvent
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type FolderMovedEvent struct {
	BaseEvent
	ParentID uuid.UUID `json:"parentId"`
}

type FolderTrashedEvent struct {
	BaseEvent
}

type FolderRestoredEvent struct {
	BaseEvent
}

type FolderPermanentlyDeletedEvent struct {
	BaseEvent
}

type NoteDeletedEvent struct {
	BaseEvent
}

type NoteUpdatedEvent struct {
	BaseEvent
	Name          string     `json:"name"`
	Icon          string     `json:"icon"`
	Tags          []string   `json:"tags"`
	Size          uint64     `json:"size"`
	FolderID      uuid.UUID  `json:"folderId"`
	OutgoingLinks uuid.UUIDs `json:"outgoingLinks"`
}

type NoteMovedEvent struct {
	BaseEvent
	FolderID uuid.UUID `json:"folderId"`
}

type NoteTrashedEvent struct {
	BaseEvent
}

type NoteRestoredEvent struct {
	BaseEvent
}

type NotePermanentlyDeletedEvent struct {
	BaseEvent
}

type WorkspaceUpdatedEvent struct {
	BaseEvent
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WorkspaceDeletedEvent struct {
	BaseEvent
}
