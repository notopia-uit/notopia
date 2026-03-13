package domain

import (
	"context"

	"github.com/google/uuid"
)

type NoteRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Note, error)
	Save(ctx context.Context, note *Note) error
	GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]Note, error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error
}
