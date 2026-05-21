package app

import (
	"context"

	"github.com/google/uuid"
)

type CheckWorkspaceSlugExistsReadModel interface {
	CheckWorkspaceSlugExists(ctx context.Context, slug string) (bool, error)
}

type GetNoteReadModelParams struct {
	ID             uuid.UUID
	ExcludeTrashed bool
}

type GetNoteReadModel interface {
	GetNote(ctx context.Context, p *GetNoteReadModelParams) (*Note, error)
}

type GetFolderReadModelParams struct {
	ID             uuid.UUID
	ExcludeTrashed bool
}

type GetFolderReadModel interface {
	GetFolder(ctx context.Context, p *GetFolderReadModelParams) (Folder, error)
}

type GetWorkspacesReadModel interface {
	GetWorkspaces(ctx context.Context, ids []uuid.UUID) ([]Workspace, error)
}

type GetNoteGraphReadModelParams struct {
	ID    uuid.UUID
	Depth int
}

type GetNoteGraphReadModel interface {
	GetNoteGraph(ctx context.Context, p *GetNoteGraphReadModelParams) (Graph, error)
}

type GetNoteLinksReadModelParams struct {
	ID            uuid.UUID
	OutgoingLinks bool
	Backlinks     bool
}

type GetNoteLinksReadModel interface {
	GetNoteLinks(ctx context.Context, p *GetNoteLinksReadModelParams) (NoteLinkResult, error)
}

type GetWorkspaceByNoteReadModel interface {
	GetWorkspaceByNoteID(ctx context.Context, noteID uuid.UUID) (Workspace, error)
}

type WorkspaceBySlugReadModel interface {
	GetWorkspaceBySlug(ctx context.Context, slug string) (Workspace, error)
}

type GetWorkspaceGraphReadModelParams struct {
	ID            uuid.UUID
	IgnoreOrphans bool
}

type GetWorkspaceGraphReadModel interface {
	GetWorkspaceGraph(ctx context.Context, p *GetWorkspaceGraphReadModelParams) (Graph, error)
}

type GetWorkspaceTreeReadModelParams struct {
	WorkspaceID    uuid.UUID
	RootFolderID   uuid.UUID
	IncludeTrashed bool
	Depth          uint
	Sort           GetWorkspaceTreeSort
}

type GetWorkspaceTreeReadModel interface {
	GetWorkspaceTree(ctx context.Context, p *GetWorkspaceTreeReadModelParams) (WorkspaceTreeFolder, error)
}
