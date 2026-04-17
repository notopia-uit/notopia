package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event interface {
	GetID() uuid.UUID
	GetOccurredAt() time.Time
	GetAggregateID() uuid.UUID
	GetUserID() string
}

type BaseEvent struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	AggregateID uuid.UUID
	UserID      string
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

type FolderCreatedEvent struct {
	BaseEvent
	Name string
	Icon string
}

type FolderDeletedEvent struct {
	BaseEvent
}

type FolderRenamedEvent struct {
	BaseEvent
	Name string
}

type FolderIconChangedEvent struct {
	BaseEvent
	Icon string
}

type NoteCreatedEvent struct {
	BaseEvent
	Name string
	Icon string
}

type FolderMovedEvent struct {
	BaseEvent
	ParentID uuid.UUID
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

type NoteRenamedEvent struct {
	BaseEvent
	Name string
}

type NoteIconChangedEvent struct {
	BaseEvent
	Icon string
}

type NoteTagsChangedEvent struct {
	BaseEvent
	Tags []string
}

type NoteSizeChangedEvent struct {
	BaseEvent
	Size uint64
}

type NoteOutgoingLinksChangedEvent struct {
	BaseEvent
	OutgoingLinks uuid.UUIDs
}

type NoteMovedEvent struct {
	BaseEvent
	FolderID uuid.UUID
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

type WorkspaceRenamedEvent struct {
	BaseEvent
	Name string
}

type WorkspaceSlugChangedEvent struct {
	BaseEvent
	Slug string
}

type WorkspaceDeletedEvent struct {
	BaseEvent
}
