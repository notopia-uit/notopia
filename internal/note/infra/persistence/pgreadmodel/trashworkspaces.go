package pgreadmodel

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type TrashWorkspaces struct {
	queries *pgsqlc.Queries
}

var _ app.TrashWorkspacesReadModel = (*TrashWorkspaces)(nil)

func NewTrashWorkspaces(queries *pgsqlc.Queries) *TrashWorkspaces {
	return &TrashWorkspaces{queries: queries}
}

var ProvideTrashWorkspaces = NewTrashWorkspaces

func (h *TrashWorkspaces) Handle(ctx context.Context, before time.Time) ([]uuid.UUID, error) {
	ids, err := h.queries.ReadWorkspaceIDsWithTrashBefore(ctx, before)
	if err != nil {
		return nil, toErr(err)
	}
	return ids, nil
}
