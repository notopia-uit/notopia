package pg

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type CheckWorkspaceSlugExistsReadModel struct {
	queries *pgsqlc.Queries
}

var _ app.CheckWorkspaceSlugExistsReadModel = (*CheckWorkspaceSlugExistsReadModel)(nil)

func NewCheckWorkspaceSlugExistsReadModel(queries *pgsqlc.Queries) *CheckWorkspaceSlugExistsReadModel {
	return &CheckWorkspaceSlugExistsReadModel{queries: queries}
}

var ProvideCheckWorkspaceSlugExistsReadModel = NewCheckWorkspaceSlugExistsReadModel

func (h *CheckWorkspaceSlugExistsReadModel) CheckWorkspaceSlugExists(ctx context.Context, q *app.CheckWorkspaceSlugExists) (*app.CheckWorkspaceSlugExistsResult, error) {
	exists, err := h.queries.CheckSlugExists(ctx, q.Slug)
	if err != nil {
		return nil, toErr(err)
	}

	return &app.CheckWorkspaceSlugExistsResult{
		Exists: exists,
	}, nil
}
