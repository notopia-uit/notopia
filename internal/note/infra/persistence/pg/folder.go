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

type Folder struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	inTransaction bool
}

var _ domain.FolderRepo = (*Folder)(nil)

func NewFolder(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	inTransaction bool,
) *Folder {
	return &Folder{
		pgxPool:       pgxPool,
		queries:       queries,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionFolder(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
) *Folder {
	return NewFolder(pgxPool, queries, false)
}

var ProvideFolder = NewNoTransactionFolder

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Folder, errs.Error) {
	folder, err := f.queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
		ID:        id,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewFolderNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	return folderToDomain(folder), nil
}

func (f *Folder) GetMany(ctx context.Context, params *domain.FolderRepoGetManyParams) ([]*domain.Folder, errs.Error) {
	var ids *[]uuid.UUID
	if len(params.IDs()) > 0 {
		paramIDs := params.IDs()
		ids = &paramIDs
	}

	var workspaceID *uuid.UUID
	if params.WorkspaceID() != nil {
		workspaceID = params.WorkspaceID()
	}

	var trashedBy *string
	if params.TrashedBy() != nil {
		trashedByStr := params.TrashedBy().String()
		trashedBy = &trashedByStr
	}

	includeTrashed := params.IsTrashed()

	folders, err := f.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		IDs:            ids,
		WorkspaceID:    workspaceID,
		TrashedBy:      trashedBy,
		IncludeTrashed: !includeTrashed,
		ForUpdate:      params.ForUpdate(),
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	result := make([]*domain.Folder, len(folders))
	for i, folder := range folders {
		result[i] = folderToDomain(folder)
	}
	return result, nil
}

func folderToDomain(folder *pgsqlc.Folder) *domain.Folder {
	var trashed *domain.Trashed
	if folder.TrashedBy != nil && folder.TrashedAt != nil {
		trashed = domain.NewTrashed(
			domain.TrashedBy(*folder.TrashedBy),
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

func (f *Folder) Save(ctx context.Context, folder *domain.Folder) (cerr errs.Error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       f.pgxPool,
		queries:       f.queries,
		inTransaction: f.inTransaction,
	}, func(queries *pgsqlc.Queries) errs.Error {
		if err := queries.SaveFolder(ctx, &pgsqlc.SaveFolderParams{
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
	})
}

func (f *Folder) SaveMany(ctx context.Context, folders []*domain.Folder) (cerr errs.Error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       f.pgxPool,
		queries:       f.queries,
		inTransaction: f.inTransaction,
	}, func(queries *pgsqlc.Queries) errs.Error {
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
		return nil
	})
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
