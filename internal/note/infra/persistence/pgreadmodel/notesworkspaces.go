package pgreadmodel

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type NotesWorkspaceIDs struct {
	queries *pgsqlc.Queries
}

var _ app.GetNotesWorkspaceIDsReadModel = (*NotesWorkspaceIDs)(nil)

func NewNotesWorkspaceIDs(queries *pgsqlc.Queries) *NotesWorkspaceIDs {
	return &NotesWorkspaceIDs{queries: queries}
}

var ProvideNotesWorkspaceIDs = NewNotesWorkspaceIDs

func (h *NotesWorkspaceIDs) Handle(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	rows, err := h.queries.ReadGetNotesWorkspaceIDs(ctx, noteIDs)
	if err != nil {
		return nil, toErr(err)
	}
	result := make(map[uuid.UUID]uuid.UUID, len(rows))
	for _, r := range rows {
		result[r.NoteID] = r.WorkspaceID
	}
	return result, nil
}
