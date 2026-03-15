package domain

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id            uuid.UUID
	name          string
	icon          *string
	tags          []string
	size          uint
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
	return &Note{
		id:       id,
		name:     name,
		icon:     icon,
		tags:     tags,
		folderID: folderID,
	}
}

func UnmarshalNote(
	id uuid.UUID,
	name string,
	icon *string,
	tags []string,
	size uint,
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
}

func (n *Note) Icon() *string {
	return n.icon
}

func (n *Note) SetIcon(icon string) {
	n.icon = &icon
}

func (n *Note) Tags() []string {
	return n.tags
}

func (n *Note) SetTags(tags []string) {
	n.tags = tags
}

func (n *Note) Size() uint {
	return n.size
}

func (n *Note) SetSize(size uint) {
	n.size = size
}

func (n *Note) FolderID() uuid.UUID {
	return n.folderID
}

func (n *Note) MoveToFolder(folderID uuid.UUID) {
	n.folderID = folderID
}

func (n *Note) OutgoingLinks() uuid.UUIDs {
	return n.outgoingLinks
}

func (n *Note) SetOutgoingLinks(outgoingLinks uuid.UUIDs) {
	n.outgoingLinks = outgoingLinks
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

func (n *Note) Trash(trashedBy TrashedBy) {
	n.trashed = NewTrashed(trashedBy, time.Now())
}

func (n *Note) Restore() {
	n.trashed = nil
}

func (n *Note) Events() []Event {
	return n.events
}

func (n *Note) AddEvent(event Event) {
	n.events = append(n.events, event)
}
