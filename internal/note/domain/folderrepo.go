package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type FolderRepoGetManyParams struct {
	WorkspaceID  *uuid.UUID
	IDs          []uuid.UUID
	TrashedBy    *TrashedBy
	IsTrashed    *bool
	ForUpdate    bool
	ParentID     *uuid.UUID
	IsRootFolder bool
}

type FolderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Folder, errs.Error)
	GetMany(ctx context.Context, params *FolderRepoGetManyParams) ([]*Folder, errs.Error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, errs.Error)
	Save(ctx context.Context, folder *Folder) errs.Error
	SaveMany(ctx context.Context, folders []*Folder) errs.Error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, errs.Error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) errs.Error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error
}
