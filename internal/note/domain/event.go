package domain

import "github.com/google/uuid"

type EventType string

var (
	TypeUnspecified EventType = "UnspecifiedEvent"

	TypeFolderCreated            EventType = "FolderCreatedEvent"
	TypeFolderDeleted            EventType = "FolderDeletedEvent"
	TypeFolderUpdated            EventType = "FolderUpdatedEvent"
	TypeFolderMoved              EventType = "FolderMovedEvent"
	TypeFolderTrashed            EventType = "FolderTrashedEvent"
	TypeFolderRestored           EventType = "FolderRestoredEvent"
	TypeFolderPermanentlyDeleted EventType = "FolderPermanentlyDeletedEvent"

	TypeNoteCreated            EventType = "NoteCreatedEvent"
	TypeNoteDeleted            EventType = "NoteDeletedEvent"
	TypeNoteUpdated            EventType = "NoteUpdatedEvent"
	TypeNoteMoved              EventType = "NoteMovedEvent"
	TypeNoteTrashed            EventType = "NoteTrashedEvent"
	TypeNoteRestored           EventType = "NoteRestoredEvent"
	TypeNotePermanentlyDeleted EventType = "NotePermanentlyDeletedEvent"

	TypeWorkspaceUpdated EventType = "WorkspaceUpdatedEvent"
	TypeWorkspaceDeleted EventType = "WorkspaceDeletedEvent"
)

func (t EventType) String() string {
	return string(t)
}

type Event interface {
	EventType() EventType
}

func NewFromEventType(eventType string) (Event, bool) {
	var concreteEvent Event
	switch EventType(eventType) {
	case TypeFolderCreated:
		concreteEvent = &FolderCreatedEvent{}
	case TypeFolderDeleted:
		concreteEvent = &FolderDeletedEvent{}
	case TypeFolderUpdated:
		concreteEvent = &FolderUpdatedEvent{}
	case TypeFolderMoved:
		concreteEvent = &FolderMovedEvent{}
	case TypeFolderTrashed:
		concreteEvent = &FolderTrashedEvent{}
	case TypeFolderRestored:
		concreteEvent = &FolderRestoredEvent{}
	case TypeFolderPermanentlyDeleted:
		concreteEvent = &FolderPermanentlyDeletedEvent{}
	case TypeNoteCreated:
		concreteEvent = &NoteCreatedEvent{}
	case TypeNoteDeleted:
		concreteEvent = &NoteDeletedEvent{}
	case TypeNoteUpdated:
		concreteEvent = &NoteUpdatedEvent{}
	case TypeNoteMoved:
		concreteEvent = &NoteMovedEvent{}
	case TypeNoteTrashed:
		concreteEvent = &NoteTrashedEvent{}
	case TypeNoteRestored:
		concreteEvent = &NoteRestoredEvent{}
	case TypeNotePermanentlyDeleted:
		concreteEvent = &NotePermanentlyDeletedEvent{}
	case TypeWorkspaceUpdated:
		concreteEvent = &WorkspaceUpdatedEvent{}
	case TypeWorkspaceDeleted:
		concreteEvent = &WorkspaceDeletedEvent{}
	}
	return concreteEvent, concreteEvent != nil
}

type FolderCreatedEvent struct {
	Id   uuid.UUID
	Name string
	Icon *string
}

var _ Event = (*FolderCreatedEvent)(nil)

func (e FolderCreatedEvent) EventType() EventType {
	return TypeFolderCreated
}

type FolderDeletedEvent struct {
	Id uuid.UUID
}

var _ Event = (*FolderDeletedEvent)(nil)

func (e FolderDeletedEvent) EventType() EventType {
	return TypeFolderDeleted
}

type FolderUpdatedEvent struct {
	ID          uuid.UUID
	Name        string
	Icon        *string
	WorkspaceID uuid.UUID
	ParentID    *uuid.UUID
}

var _ Event = (*FolderUpdatedEvent)(nil)

func (e FolderUpdatedEvent) EventType() EventType {
	return TypeFolderUpdated
}

type NoteCreatedEvent struct {
	Id   uuid.UUID
	Name string
	Icon *string
}

type FolderMovedEvent struct {
	Id       uuid.UUID
	ParentId uuid.UUID
}

var _ Event = (*FolderMovedEvent)(nil)

func (e FolderMovedEvent) EventType() EventType {
	return TypeFolderMoved
}

type FolderTrashedEvent struct {
	Id uuid.UUID
}

var _ Event = (*FolderTrashedEvent)(nil)

func (e FolderTrashedEvent) EventType() EventType {
	return TypeFolderTrashed
}

type FolderRestoredEvent struct {
	Id uuid.UUID
}

var _ Event = (*FolderRestoredEvent)(nil)

func (e FolderRestoredEvent) EventType() EventType {
	return TypeFolderRestored
}

type FolderPermanentlyDeletedEvent struct {
	Id uuid.UUID
}

var _ Event = (*FolderPermanentlyDeletedEvent)(nil)

func (e FolderPermanentlyDeletedEvent) EventType() EventType {
	return TypeFolderPermanentlyDeleted
}

var _ Event = (*NoteCreatedEvent)(nil)

func (e NoteCreatedEvent) EventType() EventType {
	return TypeNoteCreated
}

type NoteDeletedEvent struct {
	Id uuid.UUID
}

var _ Event = (*NoteDeletedEvent)(nil)

func (e NoteDeletedEvent) EventType() EventType {
	return TypeNoteDeleted
}

type NoteUpdatedEvent struct {
	ID            uuid.UUID
	Name          string
	Icon          *string
	Tags          []string
	Size          uint
	FolderID      uuid.UUID
	OutgoingLinks uuid.UUIDs
}

var _ Event = (*NoteUpdatedEvent)(nil)

func (e NoteUpdatedEvent) EventType() EventType {
	return TypeNoteUpdated
}

type NoteMovedEvent struct {
	Id       uuid.UUID
	FolderId uuid.UUID
}

var _ Event = (*NoteMovedEvent)(nil)

func (e NoteMovedEvent) EventType() EventType {
	return TypeNoteMoved
}

type NoteTrashedEvent struct {
	Id uuid.UUID
}

var _ Event = (*NoteTrashedEvent)(nil)

func (e NoteTrashedEvent) EventType() EventType {
	return TypeNoteTrashed
}

type NoteRestoredEvent struct {
	Id uuid.UUID
}

var _ Event = (*NoteRestoredEvent)(nil)

func (e NoteRestoredEvent) EventType() EventType {
	return TypeNoteRestored
}

type NotePermanentlyDeletedEvent struct {
	Id uuid.UUID
}

var _ Event = (*NotePermanentlyDeletedEvent)(nil)

func (e NotePermanentlyDeletedEvent) EventType() EventType {
	return TypeNotePermanentlyDeleted
}

type WorkspaceUpdatedEvent struct {
	Id   uuid.UUID
	Name string
	Slug string
}

var _ Event = (*WorkspaceUpdatedEvent)(nil)

func (e WorkspaceUpdatedEvent) EventType() EventType {
	return TypeWorkspaceUpdated
}

type WorkspaceDeletedEvent struct {
	Id uuid.UUID
}

var _ Event = (*WorkspaceDeletedEvent)(nil)

func (e WorkspaceDeletedEvent) EventType() EventType {
	return TypeWorkspaceDeleted
}
