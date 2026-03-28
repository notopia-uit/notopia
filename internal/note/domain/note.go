package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type Note struct {
	id            uuid.UUID
	name          string
	icon          *string
	tags          []string
	size          uint64
	folderID      uuid.UUID
	outgoingLinks uuid.UUIDs
	trashed       *Trashed

	events []Event
}

func NewNote(
	id uuid.UUID,
	name string,
	icon *string,
	tags []string,
	folderID uuid.UUID,
) *Note {
	if name == "" {
		name = "Untitled Note"
	}
	return &Note{
		id:            id,
		name:          name,
		icon:          icon,
		tags:          tags,
		folderID:      folderID,
		size:          0,
		outgoingLinks: []uuid.UUID{},
		trashed:       nil,

		events: []Event{},
	}
}

func UnmarshalNote(
	id uuid.UUID,
	name string,
	icon *string,
	tags []string,
	size uint64,
	folderID uuid.UUID,
	outgoingLinks uuid.UUIDs,
	trashed *Trashed,
) *Note {
	return &Note{
		id:            id,
		name:          name,
		icon:          icon,
		tags:          tags,
		size:          size,
		folderID:      folderID,
		outgoingLinks: outgoingLinks,
		trashed:       trashed,

		events: []Event{},
	}
}

func (n *Note) ID() uuid.UUID {
	return n.id
}

func (n *Note) Name() string {
	return n.name
}

func (n *Note) Rename(name string) {
	n.name = name
	n.AddEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(),
		AggregateID:   n.id,
		Name:          n.name,
		Icon:          n.icon,
		Tags:          n.tags,
		Size:          n.size,
		FolderID:      n.folderID,
		OutgoingLinks: n.outgoingLinks,
	})
}

func (n *Note) Icon() *string {
	return n.icon
}

func (n *Note) SetIcon(icon string) {
	n.icon = &icon
	n.AddEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(),
		AggregateID:   n.id,
		Name:          n.name,
		Icon:          n.icon,
		Tags:          n.tags,
		Size:          n.size,
		FolderID:      n.folderID,
		OutgoingLinks: n.outgoingLinks,
	})
}

func (n *Note) Tags() []string {
	return n.tags
}

func (n *Note) SetTags(tags []string) {
	n.tags = tags
	n.AddEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(),
		AggregateID:   n.id,
		Name:          n.name,
		Icon:          n.icon,
		Tags:          n.tags,
		Size:          n.size,
		FolderID:      n.folderID,
		OutgoingLinks: n.outgoingLinks,
	})
}

func (n *Note) Size() uint64 {
	return n.size
}

func (n *Note) SetSize(size uint64) {
	n.size = size
	n.AddEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(),
		AggregateID:   n.id,
		Name:          n.name,
		Icon:          n.icon,
		Tags:          n.tags,
		Size:          n.size,
		FolderID:      n.folderID,
		OutgoingLinks: n.outgoingLinks,
	})
}

func (n *Note) FolderID() uuid.UUID {
	return n.folderID
}

func (n *Note) MoveToFolder(folderID uuid.UUID) {
	n.folderID = folderID
	n.AddEvent(&NoteMovedEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: n.id,
		FolderID:    n.folderID,
	})
}

func (n *Note) OutgoingLinks() uuid.UUIDs {
	return n.outgoingLinks
}

func (n *Note) SetOutgoingLinks(outgoingLinks uuid.UUIDs) {
	n.outgoingLinks = outgoingLinks
	n.AddEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(),
		AggregateID:   n.id,
		Name:          n.name,
		Icon:          n.icon,
		Tags:          n.tags,
		Size:          n.size,
		FolderID:      n.folderID,
		OutgoingLinks: n.outgoingLinks,
	})
}

func (n *Note) IsTrashed() bool {
	return n.trashed != nil
}

func (n *Note) TrashedBy() *TrashedBy {
	if n.trashed == nil {
		return nil
	}
	return &n.trashed.by
}

func (n *Note) TrashedByString() *string {
	if n.trashed == nil {
		return nil
	}
	return new(n.trashed.by.String())
}

func (n *Note) TrashedAt() *time.Time {
	if n.trashed == nil {
		return nil
	}
	return &n.trashed.at
}

func (n *Note) Trash(trashedBy TrashedBy) errs.Error {
	if n.trashed != nil {
		return errs.NewNoteAlreadyTrashed(n.id)
	}
	n.trashed = NewTrashed(trashedBy, time.Now())
	n.AddEvent(&NoteTrashedEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: n.id,
	})
	return nil
}

func (n *Note) Restore() {
	n.trashed = nil
	n.AddEvent(&NoteRestoredEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: n.id,
	})
}

func (n *Note) AddEvent(event Event) {
	n.events = append(n.events, event)
}

func (n *Note) PopEvents() []Event {
	events := n.events
	n.events = []Event{}
	return events
}
