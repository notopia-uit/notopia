package domain

import "github.com/google/uuid"

type WorkspaceEventType string

var (
	WorkspaceEventTypeUnspecified   WorkspaceEventType = "UnspecifiedEvent"
	WorkspaceEventTypeFolderCreated WorkspaceEventType = "FolderCreatedEvent"
	WorkspaceEventTypeFolderDeleted WorkspaceEventType = "FolderDeletedEvent"
	WorkspaceEventTypeFolderUpdated WorkspaceEventType = "FolderUpdatedEvent"

	WorkspaceEventTypeNoteCreated WorkspaceEventType = "NoteCreatedEvent"
	WorkspaceEventTypeNoteDeleted WorkspaceEventType = "NoteDeletedEvent"
	WorkspaceEventTypeNoteUpdated WorkspaceEventType = "NoteUpdatedEvent"

	WorkspaceEventTypeWorkspaceUpdated WorkspaceEventType = "WorkspaceUpdatedEvent"
)

type WorkspaceEvent interface {
	EventType() WorkspaceEventType
}

func NewFromEventType(eventType string) (WorkspaceEvent, bool) {
	var concreteEvent WorkspaceEvent
	switch WorkspaceEventType(eventType) {

	case WorkspaceEventTypeFolderCreated:
		concreteEvent = &FolderCreatedEvent{}

	case WorkspaceEventTypeFolderDeleted:
		concreteEvent = &FolderDeletedEvent{}

	case WorkspaceEventTypeFolderUpdated:
		concreteEvent = &FolderUpdatedEvent{}

	case WorkspaceEventTypeNoteCreated:
		concreteEvent = &NoteCreatedEvent{}

	case WorkspaceEventTypeNoteDeleted:
		concreteEvent = &NoteDeletedEvent{}

	case WorkspaceEventTypeNoteUpdated:
		concreteEvent = &NoteUpdatedEvent{}

	case WorkspaceEventTypeWorkspaceUpdated:
		concreteEvent = &WorkspaceUpdatedEvent{}

	}
	return concreteEvent, concreteEvent != nil
}

type FolderCreatedEvent struct {
	Id   uuid.UUID
	Name string
}

var _ WorkspaceEvent = (*FolderCreatedEvent)(nil)

func (e FolderCreatedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeFolderCreated
}

type FolderDeletedEvent struct {
	Id uuid.UUID
}

var _ WorkspaceEvent = (*FolderDeletedEvent)(nil)

func (e FolderDeletedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeFolderDeleted
}

type FolderUpdatedEvent Folder

var _ WorkspaceEvent = (*FolderUpdatedEvent)(nil)

func (e FolderUpdatedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeFolderUpdated
}

type NoteCreatedEvent struct {
	Id   uuid.UUID
	Name string
}

var _ WorkspaceEvent = (*NoteCreatedEvent)(nil)

func (e NoteCreatedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeNoteCreated
}

type NoteDeletedEvent struct {
	Id uuid.UUID
}

var _ WorkspaceEvent = (*NoteDeletedEvent)(nil)

func (e NoteDeletedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeNoteDeleted
}

type NoteUpdatedEvent Note

var _ WorkspaceEvent = (*NoteUpdatedEvent)(nil)

func (e NoteUpdatedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeNoteUpdated
}

type WorkspaceUpdatedEvent Workspace

var _ WorkspaceEvent = (*WorkspaceUpdatedEvent)(nil)

func (e WorkspaceUpdatedEvent) EventType() WorkspaceEventType {
	return WorkspaceEventTypeWorkspaceUpdated
}
