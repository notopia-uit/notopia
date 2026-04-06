package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type FolderRepo struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	publisher     *Publisher // This is nil when not in transaction, because we will provide it inside a transaction
	inTransaction bool
}

var _ domain.FolderRepo = (*FolderRepo)(nil)

func NewFolderRepo(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	publisher *Publisher,
	inTransaction bool,
) *FolderRepo {
	return &FolderRepo{
		pgxPool:       pgxPool,
		queries:       queries,
		publisher:     publisher,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionFolderRepo(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
) *FolderRepo {
	return NewFolderRepo(
		pgxPool,
		queries,
		nil,
		false,
	)
}

var ProvideFolderRepo = NewNoTransactionFolderRepo

func (f *FolderRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Folder, error) {
	folder, err := f.queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
		ID:        id,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewFolderNotFound(id, err)
		}
		return nil, toErr(err)
	}
	return folderToDomainRepo(folder), nil
}

func (f *FolderRepo) GetMany(ctx context.Context, params *domain.FolderRepoGetManyParams) ([]*domain.Folder, error) {
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
		return nil, toErr(err)
	}

	result := make([]*domain.Folder, len(folders))
	for i, folder := range folders {
		result[i] = folderToDomainRepo(folder)
	}
	return result, nil
}

func folderToDomainRepo(folder *pgsqlc.Folder) *domain.Folder {
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

func (f *FolderRepo) Save(ctx context.Context, folder *domain.Folder) (cerr error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       f.pgxPool,
		queries:       f.queries,
		inTransaction: f.inTransaction,
	}, func(params *RunInTxFnparams) error {
		if err := params.queries.SaveFolder(ctx, &pgsqlc.SaveFolderParams{
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
			return toErr(err)
		}
		return nil
	})
}

func (f *FolderRepo) SaveMany(ctx context.Context, folders []*domain.Folder) (cerr error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       f.pgxPool,
		queries:       f.queries,
		inTransaction: f.inTransaction,
	}, func(params *RunInTxFnparams) error {
		if err := params.queries.CreateTempTableFolders(ctx); err != nil {
			return fmt.Errorf("failed to create temp table for folders: %w", toErr(err))
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
		affected, err := params.queries.InsertTempFolders(ctx, saveFolderParams)
		if err != nil {
			return fmt.Errorf("failed to insert folders into temp table: %w", toErr(err))
		}
		if affected != int64(len(folders)) {
			return fmt.Errorf("not all folders were inserted into temp table (expected %d, got %d)", len(folders), affected)
		}
		if err = params.queries.SaveFromTempFolders(ctx); err != nil {
			return fmt.Errorf("failed to save folders from temp table: %w", toErr(err))
		}
		return nil
	})
}

func (f *FolderRepo) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error) {
	count, err := f.queries.CountFoldersInWorkspaceByIDs(ctx, &pgsqlc.CountFoldersInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if folders are in workspace: %w", toErr(err))
	}
	return count == int64(len(ids)), nil
}

func (f *FolderRepo) PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := f.queries.PermanentlyDeleteFolderByID(ctx, id); err != nil {
		return fmt.Errorf("failed to permanently delete folder: %w", toErr(err))
	}
	return nil
}

func (f *FolderRepo) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error {
	if err := f.queries.PermanentlyDeleteFoldersByIDs(ctx, ids); err != nil {
		return fmt.Errorf("failed to permanently delete folders: %w", toErr(err))
	}
	return nil
}

func (f *FolderRepo) GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	workspaceID, err := f.queries.GetWorkspaceIDByFolderID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get workspace id for folder: %w", toErr(err))
	}
	return workspaceID, nil
}
