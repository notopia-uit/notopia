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
	workspaceID *uuid.UUID
	ids         []uuid.UUID
	trashedBy   *TrashedBy
	isTrashed   bool
	forUpdate   bool
}

func NewFolderRepoGetManyParamsByIDs(ids []uuid.UUID) *FolderRepoGetManyParams {
	//exhaustruct:ignore
	return &FolderRepoGetManyParams{
		ids: ids,
	}
}

func NewFolderRepoGetManyParamsByWorkspaceID(workspaceID uuid.UUID) *FolderRepoGetManyParams {
	//exhaustruct:ignore
	return &FolderRepoGetManyParams{
		workspaceID: &workspaceID,
	}
}

func (p *FolderRepoGetManyParams) WithWorkspaceID(workspaceID uuid.UUID) *FolderRepoGetManyParams {
	p.workspaceID = &workspaceID
	return p
}

func (p *FolderRepoGetManyParams) WithTrashed() *FolderRepoGetManyParams {
	p.isTrashed = true
	return p
}

func (p *FolderRepoGetManyParams) WithTrashedBy(trashedBy TrashedBy) *FolderRepoGetManyParams {
	p.trashedBy = &trashedBy
	p.isTrashed = true
	return p
}

func (p *FolderRepoGetManyParams) WithForUpdate() *FolderRepoGetManyParams {
	p.forUpdate = true
	return p
}

func (p *FolderRepoGetManyParams) WorkspaceID() *uuid.UUID {
	return p.workspaceID
}

func (p *FolderRepoGetManyParams) IDs() []uuid.UUID {
	return p.ids
}

func (p *FolderRepoGetManyParams) IsTrashed() bool {
	return p.isTrashed
}

func (p *FolderRepoGetManyParams) TrashedBy() *TrashedBy {
	return p.trashedBy
}

func (p *FolderRepoGetManyParams) ForUpdate() bool {
	return p.forUpdate
}
