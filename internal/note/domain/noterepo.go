package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type NoteRepoGetManyParams struct {
	WorkspaceID *uuid.UUID
	IDs         []uuid.UUID
	TrashedBy   *TrashedBy
	IsTrashed   *bool
	ForUpdate   bool
}

type NoteRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Note, errs.Error)
	GetMany(ctx context.Context, params *NoteRepoGetManyParams) ([]*Note, errs.Error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, errs.Error)
	Save(ctx context.Context, note *Note) errs.Error
	SaveMany(ctx context.Context, notes []*Note) errs.Error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, errs.Error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) errs.Error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error
}
