package pgrepo

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

type Folder struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	publisher     Publisher // This is nil when not in transaction, because we will provide it inside a transaction
	inTransaction bool
}

var _ domain.FolderRepo = (*Folder)(nil)

func NewFolder(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	publisher Publisher,
	inTransaction bool,
) *Folder {
	return &Folder{
		pgxPool:       pgxPool,
		queries:       queries,
		publisher:     publisher,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionFolder(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
) *Folder {
	return NewFolder(
		pgxPool,
		queries,
		nil,
		false,
	)
}

var ProvideFolder = NewNoTransactionFolder

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Folder, error) {
	folder, err := f.queries.GetFolder(ctx, pgsqlc.GetFolderParams{
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

func (f *Folder) GetMany(ctx context.Context, params *domain.FolderRepoGetManyParams) ([]*domain.Folder, error) {
	var ids *[]uuid.UUID
	if len(params.IDs) > 0 {
		ids = &params.IDs
	}

	var workspaceID *uuid.UUID
	if params.WorkspaceID != uuid.Nil {
		workspaceID = &params.WorkspaceID
	}

	var trashedBy *string
	if params.TrashedBy != domain.TrashedByUnspecified {
		var ok bool
		trashedBy, ok = fromDomainTrashedBy(params.TrashedBy)
		if !ok {
			return nil, fmt.Errorf("invalid trashed by value: %v", params.TrashedBy)
		}
	}

	folders, err := f.queries.GetFolders(ctx,
		&pgsqlc.GetFoldersParams{
			IDs:         ids,
			WorkspaceID: workspaceID,
			TrashedBy:   trashedBy,
			TrashedOnly: params.TrashOnly,
			ForUpdate:   params.ForUpdate,
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
	var icon string
	if folder.Icon != nil {
		icon = *folder.Icon
	}
	var trashed *domain.Trashed
	if folder.TrashedBy != nil && folder.TrashedAt != nil {
		trashed = domain.NewTrashed(
			toDomainTrashedBy(*folder.TrashedBy),
			*folder.TrashedAt,
		)
	}
	var parentID uuid.UUID
	if folder.ParentID != nil {
		parentID = *folder.ParentID
	}
	return domain.UnmarshalFolder(
		folder.ID,
		folder.Name,
		icon,
		folder.WorkspaceID,
		domain.NewFolderHierarchy(parentID),
		trashed,
		false,
	)
}

func (f *Folder) Save(ctx context.Context, folder *domain.Folder) (cerr error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       f.pgxPool,
		queries:       f.queries,
		publisher:     f.publisher,
		inTransaction: f.inTransaction,
	}, func(params *RunInTxFnparams) error {
		queries := params.queries
		if folder.Deleted() {
			if err := queries.PermanentlyDeleteFolderByID(ctx, folder.ID()); err != nil {
				return toErr(err)
			}
		} else {
			var icon *string
			if folder.Icon() != "" {
				icon = new(folder.Icon())
			}
			var parentID *uuid.UUID
			if folder.ParentID() != uuid.Nil {
				parentID = new(folder.ParentID())
			}
			var trashedBy *string
			var trashedAt *time.Time
			if folder.IsTrashed() {
				trashedBy = new(folder.TrashedBy().String())
				trashedAt = new(folder.TrashedAt())
			}
			if err := queries.SaveFolder(ctx, &pgsqlc.SaveFolderParams{
				ID:          folder.ID(),
				Name:        folder.Name(),
				Icon:        icon,
				WorkspaceID: folder.WorkspaceID(),
				ParentID:    parentID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				TrashedBy:   trashedBy,
				TrashedAt:   trashedAt,
			}); err != nil {
				return toErr(err)
			}
		}
		for _, event := range folder.PopEvents() {
			if err := params.publisher.PublishWorkspaceItem(ctx, event, folder.WorkspaceID()); err != nil {
				return fmt.Errorf("failed to publish events: %w", err)
			}
		}
		return nil
	})
}

func (f *Folder) SaveMany(ctx context.Context, folders []*domain.Folder) (cerr error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       f.pgxPool,
		queries:       f.queries,
		publisher:     f.publisher,
		inTransaction: f.inTransaction,
	}, func(params *RunInTxFnparams) error {
		var deleteIDs []uuid.UUID
		var upsertFolders []*domain.Folder

		for _, folder := range folders {
			if folder.Deleted() {
				deleteIDs = append(deleteIDs, folder.ID())
			} else {
				upsertFolders = append(upsertFolders, folder)
			}
		}

		if err := f.deleteMany(ctx, params.queries, deleteIDs); err != nil {
			return err
		}

		if err := f.upsertMany(ctx, params.queries, upsertFolders); err != nil {
			return err
		}

		for _, folder := range folders {
			for _, event := range folder.PopEvents() {
				if err := params.publisher.PublishWorkspaceItem(ctx, event, folder.WorkspaceID()); err != nil {
					return fmt.Errorf("failed to publish events: %w", err)
				}
			}
		}
		return nil
	})
}

func (f *Folder) deleteMany(ctx context.Context, queries *pgsqlc.Queries, deleteIDs []uuid.UUID) error {
	if len(deleteIDs) == 0 {
		return nil
	}
	if err := queries.PermanentlyDeleteFoldersByIDs(ctx, deleteIDs); err != nil {
		return fmt.Errorf("failed bulk delete: %w", toErr(err))
	}
	return nil
}

func (f *Folder) upsertMany(ctx context.Context, queries *pgsqlc.Queries, folders []*domain.Folder) error {
	if err := queries.CreateTempTableFolders(ctx); err != nil {
		return fmt.Errorf("failed to create temp table for folders: %w", toErr(err))
	}
	saveFolderParams := make([]*pgsqlc.InsertTempFoldersParams, len(folders))
	for i, folder := range folders {
		var icon *string
		if folder.Icon() != "" {
			icon = new(folder.Icon())
		}
		var parentID *uuid.UUID
		if folder.ParentID() != uuid.Nil {
			parentID = new(folder.ParentID())
		}
		var trashedBy *string
		var trashedAt *time.Time
		if folder.IsTrashed() {
			trashedBy = new(folder.TrashedBy().String())
			trashedAt = new(folder.TrashedAt())
		}
		saveFolderParams[i] = &pgsqlc.InsertTempFoldersParams{
			ID:          folder.ID(),
			Name:        folder.Name(),
			Icon:        icon,
			WorkspaceID: folder.WorkspaceID(),
			ParentID:    parentID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			TrashedBy:   trashedBy,
			TrashedAt:   trashedAt,
		}
	}
	affected, err := queries.InsertTempFolders(ctx, saveFolderParams)
	if err != nil {
		return fmt.Errorf("failed to insert folders into temp table: %w", toErr(err))
	}
	if affected != int64(len(folders)) {
		return fmt.Errorf("not all folders were inserted into temp table (expected %d, got %d)", len(folders), affected)
	}
	if err = queries.SaveFromTempFolders(ctx); err != nil {
		return fmt.Errorf("failed to save folders from temp table: %w", toErr(err))
	}
	return nil
}

func (f *Folder) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error) {
	count, err := f.queries.CountFoldersInWorkspaceByIDs(ctx, &pgsqlc.CountFoldersInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if folders are in workspace: %w", toErr(err))
	}
	return count == int64(len(ids)), nil
}

func (f *Folder) GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	workspaceID, err := f.queries.GetWorkspaceIDByFolderID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get workspace id for folder: %w", toErr(err))
	}
	return workspaceID, nil
}
