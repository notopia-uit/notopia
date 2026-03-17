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

type Folder struct {
	queries *pgsqlc.Queries
}

var _ domain.FolderRepo = (*Folder)(nil)

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Folder, error) {
	var result *pgsqlc.Folder
	var err error
	if forUpdate {
		result, err = f.queries.GetFolderForUpdate(ctx, &pgsqlc.GetFolderForUpdateParams{
			ID: &id,
		})
	} else {
		result, err = f.queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
			ID: &id,
		})
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrFolderNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	return folderToDomain(result), nil
}

func (f *Folder) GetByIDs(ctx context.Context, ids uuid.UUIDs, forUpdate bool) ([]domain.Folder, error) {
	var folderResults []*pgsqlc.Folder
	var err error
	if forUpdate {
		folderResults, err = f.queries.GetFoldersForUpdate(ctx, &pgsqlc.GetFoldersForUpdateParams{
			IDs: ids,
		})
	} else {
		folderResults, err = f.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
			IDs: ids,
		})
	}
	if err != nil {
		return nil, toDomainError(err)
	}
	folders := make([]domain.Folder, len(folderResults))
	for i, folder := range folderResults {
		folders[i] = *folderToDomain(folder)
	}
	return folders, nil
}

func (f *Folder) Save(ctx context.Context, folder *domain.Folder) error {
	err := f.queries.SaveFolder(ctx, &pgsqlc.SaveFolderParams{
		ID:          folder.ID(),
		Name:        folder.Name(),
		Icon:        folder.Icon(),
		WorkspaceID: folder.WorkspaceID(),
		ParentID:    folder.ParentID(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		TrashedBy:   folder.TrashedByString(),
		TrashedAt:   folder.TrashedAt(),
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) SaveMany(ctx context.Context, folders []domain.Folder) error {
	err := f.queries.CreateTempTableFolders(ctx)
	if err != nil {
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
	affected, err := f.queries.InsertTempFolders(ctx, saveFolderParams)
	if err != nil {
		return toDomainError(err)
	}
	if affected != int64(len(folders)) {
		return toDomainError(errors.New("not all folders were inserted into temp table"))
	}
	err = f.queries.SaveFromTempFolders(ctx)
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error) {
	count, err := f.queries.CountFoldersInWorkspaceByIDs(ctx, &pgsqlc.CountFoldersInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, toDomainError(err)
	}
	return count == int64(len(ids)), nil
}

func (f *Folder) GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, trashedBy domain.TrashedBy) ([]domain.Folder, error) {
	results, err := f.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &workspaceID,
		TrashedBy:   trashedBy.String(),
	})
	if err != nil {
		return nil, toDomainError(err)
	}
	folders := make([]domain.Folder, len(results))
	for i, folder := range results {
		folders[i] = *folderToDomain(folder)
	}
	return folders, nil
}

func (f *Folder) PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error {
	err := f.queries.PermanentlyDeleteFolderByID(ctx, id)
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error {
	err := f.queries.PermanentlyDeleteFoldersByIDs(ctx, ids)
	if err != nil {
		return toDomainError(err)
	}
	return nil
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
