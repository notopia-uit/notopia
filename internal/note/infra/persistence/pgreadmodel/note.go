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

func (h *Note) Handle(ctx context.Context, p *app.GetNoteReadModelParams) (*app.Note, error) {
	note, err := h.queries.ReadGetNoteByID(ctx, p.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(p.ID, err)
		}
		return nil, toErr(err)
	}

	if p.ExcludeTrashed && note.TrashedAt != nil {
		return nil, errs.NewNoteNotFound(p.ID, nil)
	}

	backlinkCount, err := h.queries.ReadCountNoteBacklinks(ctx, p.ID)
	if err != nil {
		return nil, toErr(err)
	}

	outgoingLinkCount, err := h.queries.ReadCountNoteOutgoingLinks(ctx, p.ID)
	if err != nil {
		return nil, toErr(err)
	}

	icon := ""
	if note.Icon != nil {
		icon = *note.Icon
	}

	trashed, err := toAppTrashed(note.TrashedAt, note.TrashedBy)
	if err != nil {
		return nil, err
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
