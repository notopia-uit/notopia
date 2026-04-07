package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
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
		outgoingLinks, err := h.getOutgoingLinks(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		result.OutgoingLinks = outgoingLinks
	}

	if q.Backlinks {
		backlinks, err := h.getBacklinks(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		result.Backlinks = backlinks
	}

	return &result, nil
}

func (h *GetNoteLinksReadModel) getOutgoingLinks(ctx context.Context, noteID uuid.UUID) ([]*app.NoteLink, error) {
	outgoingLinks, err := h.queries.GetNoteOutgoingLinks(ctx,
		//exhaustruct:ignore
		&pgsqlc.GetNoteOutgoingLinksParams{
			SourceID: &noteID,
		},
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	if len(outgoingLinks) == 0 {
		return []*app.NoteLink{}, nil
	}

	outgoingNotes, err := h.queries.GetNotesByParams(ctx,
		//exhaustruct:ignore
		&pgsqlc.GetNotesByParamsParams{
			IDs: &outgoingLinks,
		},
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	result := make([]*app.NoteLink, len(outgoingNotes))
	for i, linkedNote := range outgoingNotes {
		var icon string
		if linkedNote.Icon != nil {
			icon = *linkedNote.Icon
		}
		result[i] = &app.NoteLink{
			ID:   linkedNote.ID,
			Name: linkedNote.Name,
			Icon: icon,
		}
	}
	return result, nil
}

func (h *GetNoteLinksReadModel) getBacklinks(ctx context.Context, noteID uuid.UUID) ([]*app.NoteLink, error) {
	backlinks, err := h.queries.GetNoteBacklinks(ctx, noteID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	if len(backlinks) == 0 {
		return []*app.NoteLink{}, nil
	}

	backlinkNotes, err := h.queries.GetNotesByParams(ctx,
		//exhaustruct:ignore
		&pgsqlc.GetNotesByParamsParams{
			IDs: &backlinks,
		})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	result := make([]*app.NoteLink, len(backlinkNotes))
	for i, linkedNote := range backlinkNotes {
		var icon string
		if linkedNote.Icon != nil {
			icon = *linkedNote.Icon
		}
		result[i] = &app.NoteLink{
			ID:   linkedNote.ID,
			Name: linkedNote.Name,
			Icon: icon,
		}
	}
	return result, nil
}
