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
	trashedBy     *TrashedBy
	trashedAt     *time.Time
}

func NewNote(
	id uuid.UUID,
	name string,
	icon *string,
	tags []string,
	size uint,
	folderID uuid.UUID,
) *Note {
	return &Note{
		id:       id,
		name:     name,
		icon:     icon,
		tags:     tags,
		size:     size,
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
	trashedBy *TrashedBy,
	trashedAt *time.Time,
) *Note {
	return &Note{
		id:            id,
		name:          name,
		icon:          icon,
		tags:          tags,
		size:          size,
		folderID:      folderID,
		outgoingLinks: outgoingLinks,
		trashedBy:     trashedBy,
		trashedAt:     trashedAt,
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

func (n *Note) TrashedBy() *TrashedBy {
	return n.trashedBy
}

func (n *Note) TrashedAt() *time.Time {
	return n.trashedAt
}

func (n *Note) Trash(trashedBy TrashedBy) {
	n.trashedBy = &trashedBy
	n.trashedAt = new(time.Now())
}
