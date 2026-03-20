package domain

import (
	"context"

	"github.com/google/uuid"
)

type NoteRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Note, error)
	GetByIDs(ctx context.Context, ids uuid.UUIDs, forUpdate bool) ([]Note, error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	Save(ctx context.Context, note *Note) error
	SaveMany(ctx context.Context, notes []Note) error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error)
	GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]Note, error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error
}
