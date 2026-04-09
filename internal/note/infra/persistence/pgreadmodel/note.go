package pgreadmodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteReadModel = (*Note)(nil)

func GetNote(queries *pgsqlc.Queries) *Note {
	return &Note{queries: queries}
}

var ProvideNote = GetNote

func (h *Note) GetNote(ctx context.Context, q *app.GetNote) (*app.Note, error) {
	note, err := h.queries.ReadGetNoteByID(ctx, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(q.ID, err)
		}
		return nil, toErr(err)
	}

	if q.ExcludeTrashed && note.TrashedAt != nil {
		return nil, errs.NewNoteNotFound(q.ID, nil)
	}

	backlinkCount, err := h.queries.ReadCountNoteBacklinks(ctx, q.ID)
	if err != nil {
		return nil, toErr(err)
	}

	outgoingLinkCount, err := h.queries.ReadCountNoteOutgoingLinks(ctx, q.ID)
	if err != nil {
		return nil, toErr(err)
	}

	icon := ""
	if note.Icon != nil {
		icon = *note.Icon
	}

	var trashed *app.Trashed

	if note.TrashedAt != nil && note.TrashedBy != nil {
		trashedBy, err := toAppTrashedBy(note.TrashedBy)
		if err != nil {
			return nil, err
		}
		trashed = &app.Trashed{
			TrashedBy: trashedBy,
			TrashedAt: *note.TrashedAt,
		}
	}

	result := &app.Note{
		ID:                 note.ID,
		Name:               note.Name,
		Icon:               icon,
		Tags:               note.Tags,
		Size:               note.Size,
		FolderID:           note.FolderID,
		BacklinksCount:     int(backlinkCount),
		OutgoingLinksCount: int(outgoingLinkCount),
		UpdatedAt:          note.UpdatedAt,
		Trashed:            trashed,
	}

	return result, nil
}
