package domain

import (
	"context"

	"github.com/google/uuid"
)

type FolderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Folder, error)
	GetMany(ctx context.Context, params *FolderRepoGetManyParams) ([]*Folder, error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	Save(ctx context.Context, folder *Folder) error
	SaveMany(ctx context.Context, folders []*Folder) error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error)
}

type FolderRepoGetManyParams struct {
	WorkspaceID    uuid.UUID
	IDs            []uuid.UUID
	TrashedBy      TrashedBy
	NotTrashedOnly bool
	TrashOnly      bool
	ForUpdate      bool
}
