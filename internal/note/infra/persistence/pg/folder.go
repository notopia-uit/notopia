package pg

import (
	"context"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/table"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Folder struct {
	queries *pgsqlc.Queries
	db      qrm.DB
}

var _ domain.FolderRepo = (*Folder)(nil)

func NewFolder(queries *pgsqlc.Queries, db qrm.DB) *Folder {
	return &Folder{
		queries: queries,
		db:      db,
	}
}

var ProvideFolder = NewFolder

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Folder, errs.Error) {
	stmt := SELECT(table.Folders.AllColumns).
		FROM(table.Folders).
		WHERE(table.Folders.ID.EQ(UUID(id)))
	if forUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*pgsqlc.Folder
	err := stmt.QueryContext(ctx, f.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(dest) == 0 {
		return nil, errs.NewFolderNotFound(id, pgx.ErrNoRows)
	}

	result := dest[0]
	return folderToDomain(result), nil
}

func (f *Folder) GetMany(ctx context.Context, params domain.FolderRepoGetManyParams) ([]*domain.Folder, errs.Error) {
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

	var dest []*pgsqlc.Folder
	err := stmt.QueryContext(ctx, f.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}
	results := dest

	folders := make([]*domain.Folder, len(results))
	for i, folder := range results {
		folders[i] = folderToDomain(folder)
	}
	return folders, nil
}

func (f *Folder) Save(ctx context.Context, folder *domain.Folder) errs.Error {
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

func (f *Folder) SaveMany(ctx context.Context, folders []*domain.Folder) errs.Error {
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
	err := f.queries.PermanentlyDeleteFolderByID(ctx, id)
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error {
	err := f.queries.PermanentlyDeleteFoldersByIDs(ctx, ids)
	if err != nil {
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
