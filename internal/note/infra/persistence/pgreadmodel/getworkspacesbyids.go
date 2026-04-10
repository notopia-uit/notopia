package pgreadmodel

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetWorkspacesByIDs struct {
	queries *pgsqlc.Queries
}

var _ app.GetMyWorkspacesReadModel = (*GetWorkspacesByIDs)(nil)

func NewGetWorkspacesByIDs(queries *pgsqlc.Queries) *GetWorkspacesByIDs {
	return &GetWorkspacesByIDs{queries: queries}
}

var ProvideGetWorkspacesByIDs = NewGetWorkspacesByIDs

func (h *GetWorkspacesByIDs) GetWorkspacesByIDs(ctx context.Context, ids []uuid.UUID) ([]*app.Workspace, error) {
	workspaces, err := h.queries.ReadGetWorkspacesByIDs(ctx, ids)
	if err != nil {
		return nil, toErr(err)
	}
	if len(workspaces) == 0 {
		return []*app.Workspace{}, nil
	}
	result := toAppWorkspaces(workspaces)
	return result, nil
}
