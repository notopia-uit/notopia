package app

import (
	"context"

	"github.com/google/uuid"
)

type CheckWorkspaceSlugExistsReadModel interface {
	Handle(ctx context.Context, slug string) (bool, error)
}

type GetNoteReadModelParams struct {
	ID             uuid.UUID
	ExcludeTrashed bool
}

type GetNoteReadModel interface {
	Handle(ctx context.Context, p *GetNoteReadModelParams) (*Note, error)
}

type GetFolderReadModelParams struct {
	ID             uuid.UUID
	ExcludeTrashed bool
}

type GetFolderReadModel interface {
	Handle(ctx context.Context, p *GetFolderReadModelParams) (Folder, error)
}

type GetWorkspacesReadModel interface {
	Handle(ctx context.Context, ids []uuid.UUID) ([]Workspace, error)
}

type GetNoteGraphReadModelParams struct {
	ID    uuid.UUID
	Depth int
}

type GetNoteGraphReadModel interface {
	Handle(ctx context.Context, p *GetNoteGraphReadModelParams) (Graph, error)
}

type GetNoteLinksReadModelParams struct {
	ID            uuid.UUID
	OutgoingLinks bool
	Backlinks     bool
}

type GetNoteLinksReadModel interface {
	Handle(ctx context.Context, p *GetNoteLinksReadModelParams) (NoteLinkResult, error)
}

type GetWorkspaceByNoteReadModel interface {
	Handle(ctx context.Context, noteID uuid.UUID) (Workspace, error)
}

type WorkspaceBySlugReadModel interface {
	Handle(ctx context.Context, slug string) (Workspace, error)
}

type GetWorkspaceGraphReadModelParams struct {
	ID            uuid.UUID
	IgnoreOrphans bool
}

type GetWorkspaceGraphReadModel interface {
	Handle(ctx context.Context, p *GetWorkspaceGraphReadModelParams) (Graph, error)
}

type GetWorkspaceTreeReadModelParams struct {
	WorkspaceID    uuid.UUID
	RootFolderID   uuid.UUID
	IncludeTrashed bool
	Depth          uint
	Sort           GetWorkspaceTreeSort
}

type GetWorkspaceTreeReadModel interface {
	Handle(ctx context.Context, p *GetWorkspaceTreeReadModelParams) (WorkspaceTreeFolder, error)
}
