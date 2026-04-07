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

type GetNoteLinks struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteLinksReadModel = (*GetNoteLinks)(nil)

func NewGetNoteLinks(queries *pgsqlc.Queries) *GetNoteLinks {
	return &GetNoteLinks{queries: queries}
}

var ProvideGetNoteLinks = NewGetNoteLinks

func (h *GetNoteLinks) GetNoteLinks(ctx context.Context, q *app.GetNoteLinks) (*app.NoteLinkResult, error) {
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

func (h *GetNoteLinks) getOutgoingLinks(ctx context.Context, noteID uuid.UUID) ([]*app.NoteLink, error) {
	outgoingLinks, err := h.queries.ReadGetNoteOutgoingLinks(ctx, noteID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	if len(outgoingLinks) == 0 {
		return []*app.NoteLink{}, nil
	}

	outgoingNotes, err := h.queries.ReadGetNotesByIDs(ctx, outgoingLinks)
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

func (h *GetNoteLinks) getBacklinks(ctx context.Context, noteID uuid.UUID) ([]*app.NoteLink, error) {
	backlinks, err := h.queries.ReadGetNoteBacklinks(ctx, noteID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	if len(backlinks) == 0 {
		return []*app.NoteLink{}, nil
	}

	backlinkNotes, err := h.queries.ReadGetNotesByIDs(ctx, backlinks)
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
