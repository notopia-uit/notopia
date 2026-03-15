package domain

import "github.com/google/uuid"

type FolderCreatedEvent struct {
	Id   uuid.UUID
	Name string
}

type FolderDeletedEvent struct {
	Id uuid.UUID
}

type FolderUpdatedEvent Folder

type NoteCreatedEvent struct {
	Id   uuid.UUID
	Name string
}

type NoteDeletedEvent struct {
	Id uuid.UUID
}

type NoteUpdatedEvent Note

type WorkspaceUpdatedEvent Workspace
