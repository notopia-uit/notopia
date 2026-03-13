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

type Workspace struct {
	queries *pgsqlc.Queries
}

var _ domain.WorkspaceRepo = (*Workspace)(nil)

func NewWorkspaceRepo(queries *pgsqlc.Queries) domain.WorkspaceRepo {
	return &Workspace{
		queries: queries,
	}
}

func (w *Workspace) GetBySlug(slug string) (*domain.Workspace, error) {
	ctx := context.Background()
	result, err := w.queries.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceNotFound(result.ID)
		}
		return nil, toDomainError(err)
	}
	return workspaceToDomain(result), nil
}

func (w *Workspace) GetIDBySlug(slug string) (uuid.UUID, error) {
	ctx := context.Background()
	result, err := w.queries.GetWorkspaceIDBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.NewErrWorkspaceNotFound(uuid.Nil)
		}
		return uuid.Nil, toDomainError(err)
	}
	return result, nil
}

func (w *Workspace) CheckSlugExists(slug string) (bool, error) {
	ctx := context.Background()
	result, err := w.queries.CheckSlugExists(ctx, slug)
	if err != nil {
		return false, toDomainError(err)
	}
	return result, nil
}

func (w *Workspace) Save(workspace *domain.Workspace) error {
	ctx := context.Background()
	var rootFolderID pgtype.UUID
	if workspace.RootFolderID() != uuid.Nil {
		rootFolderID = pgtype.UUID{Bytes: workspace.RootFolderID(), Valid: true}
	}

	err := w.queries.SaveWorkspace(ctx, &pgsqlc.SaveWorkspaceParams{
		ID:           workspace.ID(),
		Slug:         workspace.Slug(),
		Name:         workspace.Name(),
		RootFolderID: rootFolderID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		DeletedAt:    workspace.DeletedAt(),
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func workspaceToDomain(workspace *pgsqlc.Workspace) *domain.Workspace {
	return domain.NewWorkspace(
		workspace.ID,
		workspace.Name,
		workspace.Slug,
		workspace.RootFolderID.Bytes,
	)
}
