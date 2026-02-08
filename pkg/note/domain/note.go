package domain

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id        uuid.UUID
	title     string
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewNote(
	id uuid.UUID,
	title string,
) *Note {
	now := time.Now()
	return &Note{
		id:        id,
		title:     title,
		createdAt: now,
		updatedAt: now,
	}
}

func UnmarshalNote(
	id uuid.UUID,
	title string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Note {
	return &Note{
		id:        id,
		title:     title,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

func (n *Note) ID() uuid.UUID {
	return n.id
}

func (n *Note) Title() string {
	return n.title
}

func (n *Note) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Note) UpdatedAt() time.Time {
	return n.updatedAt
}

func (n *Note) DeletedAt() *time.Time {
	return n.deletedAt
}
