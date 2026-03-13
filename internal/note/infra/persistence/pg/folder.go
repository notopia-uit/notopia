package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Folder struct {
	queries *pgsqlc.Queries
}

var _ domain.FolderRepo = (*Folder)(nil)

func NewFolderRepo(queries *pgsqlc.Queries) domain.FolderRepo {
	return &Folder{
		queries: queries,
	}
}

func (f *Folder) GetByID(ctx context.Context, id uuid.UUID) (*domain.Folder, error) {
	result, err := f.queries.GetFolder(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrFolderNotFound(id)
		}
		return nil, toDomainError(err)
	}
	return folderToDomain(result), nil
}

func (f *Folder) Save(ctx context.Context, folder *domain.Folder) error {
	var trashedBy *string
	if folder.TrashedBy() != nil {
		trashedByStr := string(*folder.TrashedBy())
		trashedBy = &trashedByStr
	}

	hierarchy := folder.FolderHierarchy()
	var parentID pgtype.UUID
	pid, _ := hierarchy.ParentID()
	if pid != nil {
		parentID = pgtype.UUID{Bytes: *pid, Valid: true}
	}

	err := f.queries.SaveFolder(ctx, &pgsqlc.SaveFolderParams{
		ID:          folder.ID(),
		Name:        folder.Name(),
		Icon:        folder.Icon(),
		WorkspaceID: folder.WorkspaceID(),
		ParentID:    parentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedBy:   trashedBy,
		DeletedAt:   folder.TrashedAt(),
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (f *Folder) GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	results, err := f.queries.GetTrashedFoldersByWorkspaceID(ctx, workspaceID)
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
	var trashedBy *domain.TrashedBy
	if folder.DeletedBy != nil {
		trashedByVal := domain.TrashedBy(*folder.DeletedBy)
		trashedBy = &trashedByVal
	}

	var parentID *uuid.UUID
	if folder.ParentID.Valid {
		parentID = (*uuid.UUID)(&folder.ParentID.Bytes)
	}

	folderHierarchy := domain.NewFolderHierarchy(parentID)

	return domain.UnmarshalFolder(
		folder.ID,
		folder.Name,
		folder.Icon,
		folder.WorkspaceID,
		*folderHierarchy,
		trashedBy,
		folder.DeletedAt,
	)
}
