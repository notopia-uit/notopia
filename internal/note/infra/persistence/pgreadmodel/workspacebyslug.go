package pgreadmodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type WorkspaceBySlug struct {
	queries *pgsqlc.Queries
}

var _ app.WorkspaceBySlugReadModel = (*WorkspaceBySlug)(nil)

func NewWorkspaceBySlug(queries *pgsqlc.Queries) *WorkspaceBySlug {
	return &WorkspaceBySlug{queries: queries}
}

var ProvideWorkspaceBySlug = NewWorkspaceBySlug

func (h *WorkspaceBySlug) GetWorkspaceBySlug(ctx context.Context, q *app.GetWorkspaceBySlug) (*app.Workspace, error) {
	workspace, err := h.queries.ReadGetWorkspaceBySlug(ctx, q.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceBySlugNotFound(q.Slug, err)
		}
		return nil, toErr(err)
	}

	return &app.Workspace{
		ID:   workspace.ID,
		Slug: workspace.Slug,
		Name: workspace.Name,
	}, nil
}
