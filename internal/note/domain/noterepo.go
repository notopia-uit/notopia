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
