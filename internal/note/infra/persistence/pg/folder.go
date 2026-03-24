package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/model"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/table"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Folder struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	db            qrm.DB
	inTransaction bool
}

var _ domain.FolderRepo = (*Folder)(nil)

func NewFolder(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	db qrm.DB,
	inTransaction bool,
) *Folder {
	return &Folder{
		pgxPool:       pgxPool,
		queries:       queries,
		db:            db,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionFolder(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	db qrm.DB,
) *Folder {
	return NewFolder(pgxPool, queries, db, false)
}

var ProvideFolder = NewNoTransactionFolder

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Folder, errs.Error) {
	stmt := SELECT(table.Folders.AllColumns).
		FROM(table.Folders).
		WHERE(table.Folders.ID.EQ(UUID(id)))
	if forUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest *model.Folders
	err := stmt.QueryContext(ctx, f.db, dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, errs.NewFolderNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	return folderToDomain(dest), nil
}

func (f *Folder) GetMany(ctx context.Context, params *domain.FolderRepoGetManyParams) ([]*domain.Folder, errs.Error) {
	condition := Bool(true)
	if params.WorkspaceID != nil {
		condition = condition.AND(table.Folders.WorkspaceID.EQ(UUID(params.WorkspaceID)))
	}
	if len(params.IDs) > 0 {
		var idExprs []Expression
		for _, id := range params.IDs {
			idExprs = append(idExprs, UUID(id))
		}
		condition = condition.AND(table.Folders.ID.IN(idExprs...))
	}
	if params.TrashedBy != nil {
		condition = condition.AND(table.Folders.TrashedBy.EQ(String(params.TrashedBy.String())))
	}
	if params.IsTrashed != nil {
		condition = condition.AND(table.Folders.TrashedAt.IS_NULL())
	}
	if params.ParentID != nil {
		condition = condition.AND(table.Folders.ParentID.EQ(UUID(params.ParentID)))
	} else if params.IsRootFolder {
		condition = condition.AND(table.Folders.ParentID.IS_NULL())
	}

	stmt := SELECT(table.Folders.AllColumns).
		FROM(table.Folders).
		WHERE(condition)
	if params.ForUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*model.Folders
	if err := stmt.QueryContext(ctx, f.db, &dest); err != nil {
		return nil, toDomainError(err)
	}
	folders := make([]*domain.Folder, len(dest))
	for i, folder := range dest {
		folders[i] = folderToDomain(folder)
	}
	return folders, nil
}

func folderToDomain(folder *model.Folders) *domain.Folder {
	var trashed *domain.Trashed
	if folder.TrashedBy != nil && folder.TrashedAt != nil {
		trashed = domain.NewTrashed(
			domain.TrashedBy(folder.TrashedBy.String()),
			*folder.TrashedAt,
		)
	}
	return domain.UnmarshalFolder(
		folder.ID,
		folder.Name,
		folder.Icon,
		folder.WorkspaceID,
		*domain.NewFolderHierarchy(folder.ParentID),
		trashed,
	)
}

func (f *Folder) Save(ctx context.Context, folder *domain.Folder) errs.Error {
	if err := f.queries.SaveFolder(ctx, &pgsqlc.SaveFolderParams{
		ID:          folder.ID(),
		Name:        folder.Name(),
		Icon:        folder.Icon(),
		WorkspaceID: folder.WorkspaceID(),
		ParentID:    folder.ParentID(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		TrashedBy:   folder.TrashedByString(),
		TrashedAt:   folder.TrashedAt(),
	}); err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) SaveMany(ctx context.Context, folders []*domain.Folder) (cerr errs.Error) {
	var queries *pgsqlc.Queries
	var tx pgx.Tx
	var err error
	if !f.inTransaction {
		tx, err = f.pgxPool.Begin(ctx)
		if err != nil {
			return toDomainError(err)
		}
		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				cerr = errs.NewPersistenceInternal("failed to rollback transaction", fmt.Errorf("%w: %v", cerr, err))
			}
		}()
		queries = f.queries.WithTx(tx)
	} else {
		queries = f.queries
	}
	if err := queries.CreateTempTableFolders(ctx); err != nil {
		return toDomainError(err)
	}
	saveFolderParams := make([]*pgsqlc.InsertTempFoldersParams, len(folders))
	for i, folder := range folders {
		saveFolderParams[i] = &pgsqlc.InsertTempFoldersParams{
			ID:          folder.ID(),
			Name:        folder.Name(),
			Icon:        folder.Icon(),
			WorkspaceID: folder.WorkspaceID(),
			ParentID:    folder.ParentID(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			TrashedBy:   folder.TrashedByString(),
			TrashedAt:   folder.TrashedAt(),
		}
	}
	affected, err := queries.InsertTempFolders(ctx, saveFolderParams)
	if err != nil {
		return toDomainError(err)
	}
	if affected != int64(len(folders)) {
		return toDomainError(errors.New("not all folders were inserted into temp table"))
	}
	if err = queries.SaveFromTempFolders(ctx); err != nil {
		return toDomainError(err)
	}
	if !f.inTransaction {
		if err := tx.Commit(ctx); err != nil {
			return errs.NewPersistenceInternal("failed to commit transaction", err)
		}
	}
	return nil
}

func (f *Folder) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, errs.Error) {
	count, err := f.queries.CountFoldersInWorkspaceByIDs(ctx, &pgsqlc.CountFoldersInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, toDomainError(err)
	}
	return count == int64(len(ids)), nil
}

func (f *Folder) PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) errs.Error {
	if err := f.queries.PermanentlyDeleteFolderByID(ctx, id); err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error {
	if err := f.queries.PermanentlyDeleteFoldersByIDs(ctx, ids); err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, errs.Error) {
	workspaceID, err := f.queries.GetWorkspaceIDByFolderID(ctx, id)
	if err != nil {
		return uuid.Nil, toDomainError(err)
	}
	return workspaceID, nil
}
