package domain

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id        uuid.UUID
	title     string
	deletedAt *time.Time
}

func NewNote(
	id uuid.UUID,
	title string,
) *Note {
	return &Note{
		id:    id,
		title: title,
	}
}

func UnmarshalNote(
	id uuid.UUID,
	title string,
	deletedAt *time.Time,
) *Note {
	return &Note{
		id:        id,
		title:     title,
		deletedAt: deletedAt,
	}
}

func (n *Note) ID() uuid.UUID {
	return n.id
}

func (n *Note) Title() string {
	return n.title
}

func (n *Note) DeletedAt() *time.Time {
	return n.deletedAt
}
