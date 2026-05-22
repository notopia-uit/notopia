package domain

import (
	"context"

	"github.com/google/uuid"
)

type FolderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Folder, error)
	GetMany(ctx context.Context, params *FolderRepoGetManyParams) ([]*Folder, error)
	GetRecursiveChildren(ctx context.Context, parms *FolderRepoGetRecursiveChildrenParams) ([]*Folder, error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	Save(ctx context.Context, folder *Folder) error
	SaveMany(ctx context.Context, folders []*Folder) error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error)
	GetParentIDs(ctx context.Context, id uuid.UUID, forUpdate bool) ([]uuid.UUID, error)
	CheckExists(ctx context.Context, id uuid.UUID) (bool, error)
}

type FolderRepoGetManyParams struct {
	WorkspaceID uuid.UUID   // Emptyable
	IDs         []uuid.UUID // Nilable
	TrashedBy   TrashedBy   // Unspecified-able
	TrashOnly   bool
	ForUpdate   bool
}

type FolderRepoGetRecursiveChildrenParams struct {
	ID          uuid.UUID
	IncludeRoot bool
	ForUpdate   bool
}
