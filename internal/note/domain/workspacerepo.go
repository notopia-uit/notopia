package domain

import (
	"context"

	"github.com/google/uuid"
)

type WorkspaceRepo interface {
	GetBySlug(ctx context.Context, slug string) (*Workspace, error)
	GetIDBySlug(ctx context.Context, slug string) (*uuid.UUID, error)
	CheckSlugExists(ctx context.Context, slug string) (bool, error)
	Save(ctx context.Context, workspace *Workspace) error
}
