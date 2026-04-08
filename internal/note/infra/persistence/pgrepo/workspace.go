package pgrepo

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
	publisher     Publisher
	runInTx       *RunInTx
	inTransaction bool
}

var _ domain.WorkspaceRepo = (*Workspace)(nil)

func NewWorkspace(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	publisher Publisher,
	runInTx *RunInTx,
	inTransaction bool,
) *Workspace {
	return &Workspace{
		pgxPool:       pgxPool,
		queries:       queries,
		publisher:     publisher,
		runInTx:       runInTx,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionWorkspace(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	runInTx *RunInTx,
) *Workspace {
	return NewWorkspace(pgxPool, queries, nil, runInTx, false)
}

var ProvideWorkspace = NewNoTransactionWorkspace

func (w *Workspace) GetBySlug(ctx context.Context, slug string, forUpdate bool) (*domain.Workspace, error) {
	workspace, err := w.queries.GetWorkspaceBySlug(ctx, pgsqlc.GetWorkspaceBySlugParams{
		Slug:      slug,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceBySlugNotFound(slug, err)
		}
		return nil, toErr(err)
	}

	folders, err := w.queries.GetFoldersByWorkspaceID(ctx, pgsqlc.GetFoldersByWorkspaceIDParams{
		WorkspaceID: workspace.ID,
		ForUpdate:   forUpdate,
	})
	if err != nil {
		return nil, toErr(err)
	}

	if len(folders) == 0 {
		return nil, errs.NewWorkspaceRootFolderNotFound(workspace.ID, pgx.ErrNoRows)
	}

	return workspaceToDomainRepo(workspace, folders[0].ID)
}

func (w *Workspace) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Workspace, error) {
	workspace, err := w.queries.GetWorkspaceByID(ctx, pgsqlc.GetWorkspaceByIDParams{
		ID:        id,
		ForUpdate: forUpdate,
	},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceNotFound(id, err)
		}
		return nil, toErr(err)
	}

	folders, err := w.queries.GetFoldersByWorkspaceID(ctx, pgsqlc.GetFoldersByWorkspaceIDParams{
		WorkspaceID: workspace.ID,
		ForUpdate:   forUpdate,
	})
	if err != nil {
		return nil, toErr(err)
	}

	if len(folders) == 0 {
		return nil, errs.NewWorkspaceRootFolderNotFound(id, pgx.ErrNoRows)
	}

	return workspaceToDomainRepo(workspace, folders[0].ID)
}

func (w *Workspace) GetIDBySlug(ctx context.Context, slug string) (*uuid.UUID, error) {
	result, err := w.queries.GetWorkspaceIDBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceBySlugNotFound(slug, err)
		}
		return nil, toErr(err)
	}
	return &result, nil
}

func (w *Workspace) CheckSlugExists(ctx context.Context, slug string) (bool, error) {
	result, err := w.queries.CheckSlugExists(ctx, slug)
	if err != nil {
		return false, toErr(err)
	}
	return result, nil
}

func (w *Workspace) Save(ctx context.Context, workspace *domain.Workspace) (cerr error) {
	return w.runInTx.Execute(ctx, &runInTxParams{
		pgxPool:       w.pgxPool,
		queries:       w.queries,
		publisher:     w.publisher,
		inTransaction: w.inTransaction,
	}, func(params *RunInTxFnparams) error {
		err := params.queries.SaveWorkspace(ctx, &pgsqlc.SaveWorkspaceParams{
			ID:        workspace.ID(),
			Slug:      workspace.Slug(),
			Name:      workspace.Name(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: workspace.DeletedAt(),
		})
		if err != nil {
			return toErr(err)
		}
		for _, event := range workspace.PopEvents() {
			if err := params.publisher.Publish(ctx, event); err != nil {
				return errs.NewPersistenceInternal("failed to publish events", err)
			}
		}
		return nil
	})
}

func workspaceToDomainRepo(workspace *pgsqlc.Workspace, rootFolderID uuid.UUID) (*domain.Workspace, error) {
	return domain.NewWorkspace(
		workspace.ID,
		workspace.Name,
		workspace.Slug,
		rootFolderID,
	)
}
