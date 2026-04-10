package pgreadmodel

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type CheckWorkspaceSlugExists struct {
	queries *pgsqlc.Queries
}

var _ app.CheckWorkspaceSlugExistsReadModel = (*CheckWorkspaceSlugExists)(nil)

func NewCheckWorkspaceSlugExists(queries *pgsqlc.Queries) *CheckWorkspaceSlugExists {
	return &CheckWorkspaceSlugExists{queries: queries}
}

var ProvideCheckWorkspaceSlugExists = NewCheckWorkspaceSlugExists

func (h *CheckWorkspaceSlugExists) CheckWorkspaceSlugExists(ctx context.Context, q *app.CheckWorkspaceSlugExists) (bool, error) {
	exists, err := h.queries.CheckSlugExists(ctx, q.Slug)
	if err != nil {
		return false, toErr(err)
	}

	return exists, nil
}
