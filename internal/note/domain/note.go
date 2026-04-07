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
	deleted       bool

	events []Event
}

func NewNote(
	id uuid.UUID,
	name string,
	icon *string,
	folderID uuid.UUID,
) *Note {
	if name == "" {
		name = "Untitled Note"
	}
	return &Note{
		id:            id,
		name:          name,
		icon:          icon,
		tags:          []string{},
		folderID:      folderID,
		size:          0,
		outgoingLinks: []uuid.UUID{},
		trashed:       nil,
		deleted:       false,

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
	deleted bool,
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
		deleted:       deleted,

		events: []Event{},
	}
}

func (n *Note) ID() uuid.UUID {
	return n.id
}

func (n *Note) Name() string {
	return n.name
}

func (n *Note) Rename(name string, userID string) {
	n.name = name
	n.addEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(n.id, userID),
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

func (n *Note) SetIcon(icon string, userID string) {
	n.icon = &icon
	n.addEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(n.id, userID),
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

func (n *Note) SetTags(tags []string, userID string) {
	n.tags = tags
	n.addEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(n.id, userID),
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

func (n *Note) SetSize(size uint64, userID string) {
	n.size = size
	n.addEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(n.id, userID),
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

func (n *Note) MoveToFolder(folderID uuid.UUID, userID string) {
	n.folderID = folderID
	n.addEvent(&NoteMovedEvent{
		BaseEvent: *NewBaseEvent(n.id, userID),
		FolderID:  n.folderID,
	})
}

func (n *Note) OutgoingLinks() uuid.UUIDs {
	return n.outgoingLinks
}

func (n *Note) SetOutgoingLinks(outgoingLinks uuid.UUIDs, userID string) {
	n.outgoingLinks = outgoingLinks
	n.addEvent(&NoteUpdatedEvent{
		BaseEvent:     *NewBaseEvent(n.id, userID),
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

func (n *Note) Trash(trashedBy TrashedBy, userID string) error {
	if n.trashed != nil {
		return errs.NewNoteAlreadyTrashed(n.id)
	}
	n.trashed = NewTrashed(trashedBy, time.Now())
	n.addEvent(&NoteTrashedEvent{
		BaseEvent: *NewBaseEvent(n.id, userID),
	})
	return nil
}

func (n *Note) Restore(userID string) {
	n.trashed = nil
	n.addEvent(&NoteRestoredEvent{
		BaseEvent: *NewBaseEvent(n.id, userID),
	})
}

func (n *Note) Deleted() bool {
	return n.deleted
}

func (n *Note) Delete(userID string) {
	n.deleted = true
	n.addEvent(&NoteDeletedEvent{
		BaseEvent: *NewBaseEvent(n.id, userID),
	})
}

func (n *Note) addEvent(event Event) {
	n.events = append(n.events, event)
}

func (n *Note) PopEvents() []Event {
	events := n.events
	n.events = []Event{}
	return events
}
