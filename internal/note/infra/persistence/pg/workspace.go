package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Workspace struct {
	queries *pgsqlc.Queries
}

var _ domain.WorkspaceRepo = (*Workspace)(nil)

func NewWorkspace(queries *pgsqlc.Queries) *Workspace {
	return &Workspace{
		queries: queries,
	}
}

var ProvideWorkspace = NewWorkspace

func (w *Workspace) GetBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	workspaceResult, err := w.queries.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceNotFound(slug, err)
		}
		return nil, toDomainError(err)
	}
	rootFolderResult, err := w.queries.GetRootFolder(ctx, workspaceResult.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceRootFolderNotFound(slug, err)
		}
		return nil, toDomainError(err)
	}
	return workspaceToDomain(workspaceResult, rootFolderResult.ID), nil
}

func (w *Workspace) GetIDBySlug(ctx context.Context, slug string) (*uuid.UUID, error) {
	result, err := w.queries.GetWorkspaceIDBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceNotFound(slug, err)
		}
		return nil, toDomainError(err)
	}
	return &result, nil
}

func (w *Workspace) CheckSlugExists(ctx context.Context, slug string) (bool, error) {
	result, err := w.queries.CheckSlugExists(ctx, slug)
	return result, toDomainError(err)
}

func (w *Workspace) Save(ctx context.Context, workspace *domain.Workspace) error {
	err := w.queries.SaveWorkspace(ctx, &pgsqlc.SaveWorkspaceParams{
		ID:        workspace.ID(),
		Slug:      workspace.Slug(),
		Name:      workspace.Name(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: workspace.DeletedAt(),
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func workspaceToDomain(workspace *pgsqlc.Workspace, rootFolderID uuid.UUID) *domain.Workspace {
	return domain.NewWorkspace(
		workspace.ID,
		workspace.Name,
		workspace.Slug,
		rootFolderID,
	)
}
