package domain

import (
	"context"

	"github.com/google/uuid"
)

type NoteRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Note, error)
	GetMany(ctx context.Context, params *NoteRepoGetManyParams) ([]*Note, error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	Save(ctx context.Context, note *Note) error
	SaveMany(ctx context.Context, notes []*Note) error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error)
}

type NoteRepoGetManyParams struct {
	WorkspaceID uuid.UUID
	IDs         []uuid.UUID
	TrashedBy   TrashedBy
	TrashOnly   bool
	ForUpdate   bool
}
