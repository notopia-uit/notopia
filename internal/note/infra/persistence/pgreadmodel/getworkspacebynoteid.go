package pgreadmodel

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetWorkspaceByNote struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspaceByNoteReadModel = (*GetWorkspaceByNote)(nil)

func NewGetWorkspaceByNote(queries *pgsqlc.Queries) *GetWorkspaceByNote {
	return &GetWorkspaceByNote{queries: queries}
}

var ProvideGetWorkspaceByNote = NewGetWorkspaceByNote

func (h *GetWorkspaceByNote) GetWorkspaceByNoteID(ctx context.Context, noteID uuid.UUID) (*app.Workspace, error) {
	workspaces, err := h.queries.ReadGetWorkspaceByNoteID(ctx, noteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceByNoteNotFound(noteID, err)
		}
		return nil, toErr(err)
	}
	result := toAppWorkspace(workspaces)
	return result, nil
}
