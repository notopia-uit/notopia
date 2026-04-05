package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Workspace struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	inTransaction bool
}

var _ domain.WorkspaceRepo = (*Workspace)(nil)

func NewWorkspace(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	inTransaction bool,
) *Workspace {
	return &Workspace{
		pgxPool:       pgxPool,
		queries:       queries,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionWorkspace(pgxPool *pgxpool.Pool, queries *pgsqlc.Queries) *Workspace {
	return NewWorkspace(pgxPool, queries, false)
}

var ProvideWorkspace = NewNoTransactionWorkspace

func (w *Workspace) GetBySlug(ctx context.Context, slug string, forUpdate bool) (*domain.Workspace, errs.Error) {
	workspace, err := w.queries.GetWorkspace(ctx, &pgsqlc.GetWorkspaceParams{
		Slug:      &slug,
		ID:        nil,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceBySlugNotFound(slug, err)
		}
		return nil, toDomainError(err)
	}

	folders, err := w.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID:  &workspace.ID,
		IsRootFolder: true,
		ForUpdate:    forUpdate,
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(folders) == 0 {
		return nil, errs.NewWorkspaceRootFolderNotFound(workspace.ID, pgx.ErrNoRows)
	}

	return workspaceToDomain(workspace, folders[0].ID)
}

func (w *Workspace) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Workspace, errs.Error) {
	workspace, err := w.queries.GetWorkspace(ctx, &pgsqlc.GetWorkspaceParams{
		Slug:      nil,
		ID:        &id,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceNotFound(id, err)
		}
		return nil, toDomainError(err)
	}

	folders, err := w.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID:  &workspace.ID,
		IsRootFolder: true,
		ForUpdate:    forUpdate,
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(folders) == 0 {
		return nil, errs.NewWorkspaceRootFolderNotFound(id, pgx.ErrNoRows)
	}

	return workspaceToDomain(workspace, folders[0].ID)
}

func (w *Workspace) GetIDBySlug(ctx context.Context, slug string) (*uuid.UUID, errs.Error) {
	result, err := w.queries.GetWorkspaceIDBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceBySlugNotFound(slug, err)
		}
		return nil, toDomainError(err)
	}
	return &result, nil
}

func (w *Workspace) CheckSlugExists(ctx context.Context, slug string) (bool, errs.Error) {
	result, err := w.queries.CheckSlugExists(ctx, slug)
	return result, toDomainError(err)
}

func (w *Workspace) Save(ctx context.Context, workspace *domain.Workspace) (cerr errs.Error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       w.pgxPool,
		queries:       w.queries,
		inTransaction: w.inTransaction,
	}, func(queries *pgsqlc.Queries) errs.Error {
		err := queries.SaveWorkspace(ctx, &pgsqlc.SaveWorkspaceParams{
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
	})
}

func workspaceToDomain(workspace *pgsqlc.Workspace, rootFolderID uuid.UUID) (*domain.Workspace, errs.Error) {
	return domain.NewWorkspace(
		workspace.ID,
		workspace.Name,
		workspace.Slug,
		rootFolderID,
	)
}
