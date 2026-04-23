package pgreadmodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type ShowTrash struct {
	queries *pgsqlc.Queries
}

var _ app.ShowTrashReadModel = (*ShowTrash)(nil)

func NewShowTrash(queries *pgsqlc.Queries) *ShowTrash {
	return &ShowTrash{queries: queries}
}

var ProvideShowTrash = NewShowTrash

func (h *ShowTrash) ShowTrash(ctx context.Context, q *app.ShowTrash) (app.Trash, error) {
	trashedNotes, err := h.queries.ReadGetTrashedNotesByWorkspaceID(ctx, q.WorkspaceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return app.Trash{}, toErr(err)
	}

	trashedFolders, err := h.queries.ReadGetTrashedFolderByWorkspaceID(ctx, q.WorkspaceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return app.Trash{}, toErr(err)
	}

	notes := make([]app.TrashedNote, len(trashedNotes))
	for i, note := range trashedNotes {
		var icon string
		if note.Icon != nil {
			icon = *note.Icon
		}
		trashed, err := toAppTrashed(note.TrashedAt, note.TrashedBy)
		if err != nil {
			return app.Trash{}, err
		}
		notes[i] = app.TrashedNote{
			ID:      note.ID,
			Name:    note.Name,
			Icon:    icon,
			Trashed: trashed,
		}
	}

	folders := make([]app.TrashedFolder, len(trashedFolders))
	for i, folder := range trashedFolders {
		var icon string
		if folder.Icon != nil {
			icon = *folder.Icon
		}
		trashed, err := toAppTrashed(folder.TrashedAt, folder.TrashedBy)
		if err != nil {
			return app.Trash{}, err
		}
		folders[i] = app.TrashedFolder{
			ID:      folder.ID,
			Name:    folder.Name,
			Icon:    icon,
			Trashed: trashed,
		}
	}

	return app.Trash{
		Notes:   notes,
		Folders: folders,
	}, nil
}
