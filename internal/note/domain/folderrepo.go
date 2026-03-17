package domain

import (
	"context"

	"github.com/google/uuid"
)

type FolderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	GetByIDs(ctx context.Context, ids uuid.UUIDs) ([]Folder, error)
	Save(ctx context.Context, folder *Folder) error
	SaveMany(ctx context.Context, folders []Folder) error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error)
	GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, trashedBy TrashedBy) ([]Folder, error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error
}
