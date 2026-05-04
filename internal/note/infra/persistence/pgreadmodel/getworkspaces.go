package pgreadmodel

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetWorkspaces struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspacesReadModel = (*GetWorkspaces)(nil)

func NewGetWorkspaces(queries *pgsqlc.Queries) *GetWorkspaces {
	return &GetWorkspaces{queries: queries}
}

var ProvideGetWorkspaces = NewGetWorkspaces

func (h *GetWorkspaces) GetWorkspaces(ctx context.Context, ids []uuid.UUID) ([]app.Workspace, error) {
	workspaces, err := h.queries.ReadGetWorkspacesByIDs(ctx, ids)
	if err != nil {
		return nil, toErr(err)
	}
	if len(workspaces) == 0 {
		return []app.Workspace{}, nil
	}
	result := toAppWorkspaces(workspaces)
	return result, nil
}
