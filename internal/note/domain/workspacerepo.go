package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type WorkspaceRepo interface {
	GetBySlug(ctx context.Context, slug string, forUpdate bool) (*Workspace, errs.Error)
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*Workspace, errs.Error)
	GetIDBySlug(ctx context.Context, slug string) (*uuid.UUID, errs.Error)
	CheckSlugExists(ctx context.Context, slug string) (bool, errs.Error)
	Save(ctx context.Context, workspace *Workspace) errs.Error
}
