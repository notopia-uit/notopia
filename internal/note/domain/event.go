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

type Event any

type BaseEvent struct {
	ID         uuid.UUID
	OccurredAt time.Time
}

func NewBaseEvent() *BaseEvent {
	return &BaseEvent{
		ID:         uuid.New(),
		OccurredAt: time.Now(),
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

func NewEmptyFromEventType(eventType string) (Event, bool) {
	var concreteEvent Event
	switch EventType(eventType) {
	case EventTypeFolderCreated:
		//exhaustruct:ignore
		concreteEvent = &FolderCreatedEvent{}
	case EventTypeFolderDeleted:
		//exhaustruct:ignore
		concreteEvent = &FolderDeletedEvent{}
	case EventTypeFolderUpdated:
		//exhaustruct:ignore
		concreteEvent = &FolderUpdatedEvent{}
	case EventTypeFolderMoved:
		//exhaustruct:ignore
		concreteEvent = &FolderMovedEvent{}
	case EventTypeFolderTrashed:
		//exhaustruct:ignore
		concreteEvent = &FolderTrashedEvent{}
	case EventTypeFolderRestored:
		//exhaustruct:ignore
		concreteEvent = &FolderRestoredEvent{}
	case EventTypeFolderPermanentlyDeleted:
		//exhaustruct:ignore
		concreteEvent = &FolderPermanentlyDeletedEvent{}
	case EventTypeNoteCreated:
		//exhaustruct:ignore
		concreteEvent = &NoteCreatedEvent{}
	case EventTypeNoteDeleted:
		//exhaustruct:ignore
		concreteEvent = &NoteDeletedEvent{}
	case EventTypeNoteUpdated:
		//exhaustruct:ignore
		concreteEvent = &NoteUpdatedEvent{}
	case EventTypeNoteMoved:
		//exhaustruct:ignore
		concreteEvent = &NoteMovedEvent{}
	case EventTypeNoteTrashed:
		//exhaustruct:ignore
		concreteEvent = &NoteTrashedEvent{}
	case EventTypeNoteRestored:
		//exhaustruct:ignore
		concreteEvent = &NoteRestoredEvent{}
	case EventTypeNotePermanentlyDeleted:
		//exhaustruct:ignore
		concreteEvent = &NotePermanentlyDeletedEvent{}
	case EventTypeWorkspaceUpdated:
		//exhaustruct:ignore
		concreteEvent = &WorkspaceUpdatedEvent{}
	case EventTypeWorkspaceDeleted:
		//exhaustruct:ignore
		concreteEvent = &WorkspaceDeletedEvent{}
	}
	return concreteEvent, concreteEvent != nil
}

type FolderCreatedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
	Name        string
	Icon        *string
}

type FolderDeletedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type FolderUpdatedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
	Name        string
	Icon        *string
}

type NoteCreatedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
	Name        string
	Icon        *string
}

type FolderMovedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
	ParentID    uuid.UUID
}

type FolderTrashedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type FolderRestoredEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type FolderPermanentlyDeletedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type NoteDeletedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}
type NoteUpdatedEvent struct {
	BaseEvent
	AggregateID   uuid.UUID
	Name          string
	Icon          *string
	Tags          []string
	Size          uint64
	FolderID      uuid.UUID
	OutgoingLinks uuid.UUIDs
}

type NoteMovedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
	FolderID    uuid.UUID
}

type NoteTrashedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type NoteRestoredEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type NotePermanentlyDeletedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}

type WorkspaceUpdatedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
	Name        string
	Slug        string
}

type WorkspaceDeletedEvent struct {
	BaseEvent
	AggregateID uuid.UUID
}
