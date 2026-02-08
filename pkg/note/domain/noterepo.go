package domain

import (
	"context"

	"github.com/google/uuid"
)

type NoteRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Note, error)
}
