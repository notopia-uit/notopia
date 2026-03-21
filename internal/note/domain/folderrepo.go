package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type FolderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Folder, errs.Error)
	GetByIDs(ctx context.Context, ids uuid.UUIDs, forUpdate bool) ([]Folder, errs.Error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, errs.Error)
	Save(ctx context.Context, folder *Folder) errs.Error
	SaveMany(ctx context.Context, folders []Folder) errs.Error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, errs.Error)
	GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, trashedBy TrashedBy) ([]Folder, errs.Error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) errs.Error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error
}
