package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type NoteRepoGetByWorkspaceIDParams struct {
	WorkspaceID uuid.UUID
	TrashedBy   *TrashedBy
}

type NoteRepo interface {
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Note, errs.Error)
	GetByIDs(ctx context.Context, ids uuid.UUIDs, forUpdate bool) ([]*Note, errs.Error)
	GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, errs.Error)
	GetByWorkspaceID(ctx context.Context, params NoteRepoGetByWorkspaceIDParams) ([]*Note, errs.Error)
	Save(ctx context.Context, note *Note) errs.Error
	SaveMany(ctx context.Context, notes []*Note) errs.Error
	AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, errs.Error)
	PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) errs.Error
	PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error
}
