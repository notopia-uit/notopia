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

func NewFolder(queries *pgsqlc.Queries) *Folder {
	return &Folder{
		queries: queries,
	}
}

var ProvideFolder = NewFolder

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID) (*domain.Folder, error) {
	result, err := f.queries.GetFolder(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrFolderNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	return folderToDomain(result), nil
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

func (f *Folder) GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, trashedBy domain.TrashedBy) ([]domain.Folder, error) {
	results, err := f.queries.GetTrashedFoldersByWorkspaceID(ctx, &pgsqlc.GetTrashedFoldersByWorkspaceIDParams{
		WorkspaceID: workspaceID,
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
