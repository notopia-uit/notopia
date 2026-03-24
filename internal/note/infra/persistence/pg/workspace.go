package pg

import (
	"context"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/table"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Workspace struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	db            qrm.DB
	inTransaction bool
}

var _ domain.WorkspaceRepo = (*Workspace)(nil)

func NewWorkspace(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	db qrm.DB,
	inTransaction bool,
) *Workspace {
	return &Workspace{
		pgxPool:       pgxPool,
		queries:       queries,
		db:            db,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionWorkspace(pgxPool *pgxpool.Pool, queries *pgsqlc.Queries, db qrm.DB) *Workspace {
	return NewWorkspace(pgxPool, queries, db, false)
}

var ProvideWorkspace = NewNoTransactionWorkspace

func (w *Workspace) GetBySlug(ctx context.Context, slug string, forUpdate bool) (*domain.Workspace, errs.Error) {
	stmt := SELECT(table.Workspaces.AllColumns).
		FROM(table.Workspaces).
		WHERE(table.Workspaces.Slug.EQ(String(slug)))
	if forUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*pgsqlc.Workspace
	err := stmt.QueryContext(ctx, w.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(dest) == 0 {
		return nil, errs.NewWorkspaceBySlugNotFound(slug, pgx.ErrNoRows)
	}
	workspaceResult := dest[0]

	folderStmt := SELECT(table.Folders.AllColumns).
		FROM(table.Folders).
		WHERE(table.Folders.WorkspaceID.EQ(UUID(workspaceResult.ID)).AND(table.Folders.ParentID.IS_NULL()))
	if forUpdate {
		folderStmt = folderStmt.FOR(UPDATE())
	}

	var folderDest []*pgsqlc.Folder
	err = folderStmt.QueryContext(ctx, w.db, &folderDest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(folderDest) == 0 {
		return nil, errs.NewWorkspaceRootFolderNotFound(workspaceResult.ID, pgx.ErrNoRows)
	}
	rootFolderResult := folderDest[0]

	return workspaceToDomain(workspaceResult, rootFolderResult.ID)
}

func (w *Workspace) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Workspace, errs.Error) {
	stmt := SELECT(table.Workspaces.AllColumns).
		FROM(table.Workspaces).
		WHERE(table.Workspaces.ID.EQ(UUID(id)))
	if forUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*pgsqlc.Workspace
	err := stmt.QueryContext(ctx, w.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(dest) == 0 {
		return nil, errs.NewWorkspaceNotFound(id, pgx.ErrNoRows)
	}
	workspaceResult := dest[0]

	folderStmt := SELECT(table.Folders.AllColumns).
		FROM(table.Folders).
		WHERE(table.Folders.WorkspaceID.EQ(UUID(workspaceResult.ID)).AND(table.Folders.ParentID.IS_NULL()))
	if forUpdate {
		folderStmt = folderStmt.FOR(UPDATE())
	}

	var folderDest []*pgsqlc.Folder
	err = folderStmt.QueryContext(ctx, w.db, &folderDest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(folderDest) == 0 {
		return nil, errs.NewWorkspaceRootFolderNotFound(id, pgx.ErrNoRows)
	}
	rootFolderResult := folderDest[0]

	return workspaceToDomain(workspaceResult, rootFolderResult.ID)
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

func (w *Workspace) Save(ctx context.Context, workspace *domain.Workspace) errs.Error {
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

func workspaceToDomain(workspace *pgsqlc.Workspace, rootFolderID uuid.UUID) (*domain.Workspace, errs.Error) {
	return domain.NewWorkspace(
		workspace.ID,
		workspace.Name,
		workspace.Slug,
		rootFolderID,
	)
}
