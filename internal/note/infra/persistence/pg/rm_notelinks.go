package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetNoteLinksReadModel struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteLinksReadModel = (*GetNoteLinksReadModel)(nil)

func NewGetNoteLinksReadModel(queries *pgsqlc.Queries) *GetNoteLinksReadModel {
	return &GetNoteLinksReadModel{queries: queries}
}

var ProvideGetNoteLinksReadModel = NewGetNoteLinksReadModel

func (h *GetNoteLinksReadModel) GetNoteLinks(ctx context.Context, q *app.GetNoteLinks) (*app.NoteLinkResult, error) {
	_, err := h.queries.GetNoteByID(ctx,
		//exhaustruct:ignore
		pgsqlc.GetNoteByIDParams{
			ID: q.ID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(q.ID, err)
		}
		return nil, toErr(err)
	}

	result := app.NoteLinkResult{
		OutgoingLinks: []*app.NoteLink{},
		Backlinks:     []*app.NoteLink{},
	}

	if q.OutgoingLinks {
		outgoingLinks, err := h.queries.GetNoteOutgoingLinks(ctx,
			//exhaustruct:ignore
			&pgsqlc.GetNoteOutgoingLinksParams{
				SourceID: &q.ID,
			},
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toErr(err)
		}

		if len(outgoingLinks) > 0 {
			outgoingNotes, err := h.queries.GetNotesByParams(ctx,
				//exhaustruct:ignore
				&pgsqlc.GetNotesByParamsParams{
					IDs: &outgoingLinks,
				},
			)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, toErr(err)
			}
			for _, linkedNote := range outgoingNotes {
				var icon string
				if linkedNote.Icon != nil {
					icon = *linkedNote.Icon
				}
				result.OutgoingLinks = append(result.OutgoingLinks, &app.NoteLink{
					ID:   linkedNote.ID,
					Name: linkedNote.Name,
					Icon: icon,
				})
			}
		}
	}

	if q.Backlinks {
		backlinks, err := h.queries.GetNoteBacklinks(ctx, q.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toErr(err)
		}

		if len(backlinks) > 0 {
			backlinkNotes, err := h.queries.GetNotesByParams(ctx,
				//exhaustruct:ignore
				&pgsqlc.GetNotesByParamsParams{
					IDs: &backlinks,
				})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, toErr(err)
			}
			for _, linkedNote := range backlinkNotes {
				var icon string
				if linkedNote.Icon != nil {
					icon = *linkedNote.Icon
				}
				result.Backlinks = append(result.Backlinks, &app.NoteLink{
					ID:   linkedNote.ID,
					Name: linkedNote.Name,
					Icon: icon,
				})
			}
		}
	}

	return &result, nil
}
