package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type NoteRepoGetManyParams struct {
	workspaceID *uuid.UUID
	ids         []uuid.UUID
	trashedBy   *TrashedBy
	isTrashed   *bool
	forUpdate   bool
}

func NewNoteRepoGetManyParamsByIDs(ids []uuid.UUID) *NoteRepoGetManyParams {
	//exhaustruct:ignore
	return &NoteRepoGetManyParams{
		ids: ids,
	}
}

func NewNoteRepoGetManyParamsByWorkspaceID(workspaceID uuid.UUID) *NoteRepoGetManyParams {
	//exhaustruct:ignore
	return &NoteRepoGetManyParams{
		workspaceID: &workspaceID,
	}
}

func (p *NoteRepoGetManyParams) WithWorkspaceID(workspaceID uuid.UUID) *NoteRepoGetManyParams {
	p.workspaceID = &workspaceID
	return p
}

func (p *NoteRepoGetManyParams) WithIDs(ids []uuid.UUID) *NoteRepoGetManyParams {
	p.ids = ids
	return p
}

func (p *NoteRepoGetManyParams) WithTrashedBy(trashedBy TrashedBy) *NoteRepoGetManyParams {
	p.trashedBy = &trashedBy
	p.isTrashed = new(bool)
	*p.isTrashed = true
	return p
}

func (p *NoteRepoGetManyParams) WithIsTrashed(isTrashed bool) *NoteRepoGetManyParams {
	p.isTrashed = &isTrashed
	return p
}

func (p *NoteRepoGetManyParams) WithForUpdate() *NoteRepoGetManyParams {
	p.forUpdate = true
	return p
}

func (p *NoteRepoGetManyParams) WorkspaceID() *uuid.UUID {
	return p.workspaceID
}

func (p *NoteRepoGetManyParams) IDs() []uuid.UUID {
	return p.ids
}

func (p *NoteRepoGetManyParams) TrashedBy() *TrashedBy {
	return p.trashedBy
}

func (p *NoteRepoGetManyParams) IsTrashed() *bool {
	return p.isTrashed
}

func (p *NoteRepoGetManyParams) ForUpdate() bool {
	return p.forUpdate
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
