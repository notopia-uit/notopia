package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type ShowTrashReadModel struct {
	queries *pgsqlc.Queries
}

var _ app.ShowTrashReadModel = (*ShowTrashReadModel)(nil)

func NewShowTrashReadModel(queries *pgsqlc.Queries) *ShowTrashReadModel {
	return &ShowTrashReadModel{queries: queries}
}

var ProvideShowTrashReadModel = NewShowTrashReadModel

func (h *ShowTrashReadModel) ShowTrash(ctx context.Context, q *app.ShowTrash) (*app.Trash, error) {
	trashedNotes, err := h.queries.GetTrashedNotesByWorkspaceID(ctx, q.WorkspaceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	trashedByPurpose := string(app.TrashedByPurpose)
	trashedFolders, err := h.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID:    &q.WorkspaceID,
		TrashedBy:      &trashedByPurpose,
		IncludeTrashed: true,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	notes := make([]*app.TrashedNote, len(trashedNotes))
	for i, note := range trashedNotes {
		notes[i] = &app.TrashedNote{
			ID:   note.ID,
			Name: note.Name,
			Icon: note.Icon,
			Trashed: app.Trashed{
				TrashedBy: app.TrashedByPurpose,
				TrashedAt: *note.TrashedAt,
			},
		}
	}

	folders := make([]*app.TrashedFolder, len(trashedFolders))
	for i, folder := range trashedFolders {
		folders[i] = &app.TrashedFolder{
			ID:   folder.ID,
			Name: folder.Name,
			Icon: folder.Icon,
			Trashed: app.Trashed{
				TrashedBy: app.TrashedByPurpose,
				TrashedAt: *folder.TrashedAt,
			},
		}
	}

	return &app.Trash{
		Notes:   notes,
		Folders: folders,
	}, nil
}
