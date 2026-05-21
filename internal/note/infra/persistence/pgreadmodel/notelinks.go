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

type NoteLinks struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteLinksReadModel = (*NoteLinks)(nil)

func GetNoteLinks(queries *pgsqlc.Queries) *NoteLinks {
	return &NoteLinks{queries: queries}
}

var ProvideNoteLinks = GetNoteLinks

func (h *NoteLinks) Handle(ctx context.Context, p *app.GetNoteLinksReadModelParams) (app.NoteLinkResult, error) {
	_, err := h.queries.GetNoteByID(ctx,
		//exhaustruct:ignore
		pgsqlc.GetNoteByIDParams{
			ID: p.ID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.NoteLinkResult{}, errs.NewNoteNotFound(p.ID, err)
		}
		return app.NoteLinkResult{}, toErr(err)
	}

	result := app.NoteLinkResult{
		OutgoingLinks: []app.NoteLink{},
		Backlinks:     []app.NoteLink{},
	}

	if p.OutgoingLinks {
		outgoingLinks, err := h.getOutgoingLinks(ctx, p.ID)
		if err != nil {
			return app.NoteLinkResult{}, err
		}
		result.OutgoingLinks = outgoingLinks
	}

	if p.Backlinks {
		backlinks, err := h.getBacklinks(ctx, p.ID)
		if err != nil {
			return app.NoteLinkResult{}, err
		}
		result.Backlinks = backlinks
	}

	return result, nil
}

func (h *NoteLinks) getOutgoingLinks(ctx context.Context, noteID uuid.UUID) ([]app.NoteLink, error) {
	outgoingLinks, err := h.queries.ReadGetNoteOutgoingLinks(ctx, noteID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	if len(outgoingLinks) == 0 {
		return []app.NoteLink{}, nil
	}

	outgoingNotes, err := h.queries.ReadGetNotesByIDs(ctx, outgoingLinks)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	result := make([]app.NoteLink, len(outgoingNotes))
	for i, linkedNote := range outgoingNotes {
		var icon string
		if linkedNote.Icon != nil {
			icon = *linkedNote.Icon
		}
		result[i] = app.NoteLink{
			ID:   linkedNote.ID,
			Name: linkedNote.Name,
			Icon: icon,
		}
	}
	return result, nil
}

func (h *NoteLinks) getBacklinks(ctx context.Context, noteID uuid.UUID) ([]app.NoteLink, error) {
	backlinks, err := h.queries.ReadGetNoteBacklinks(ctx, noteID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	if len(backlinks) == 0 {
		return []app.NoteLink{}, nil
	}

	backlinkNotes, err := h.queries.ReadGetNotesByIDs(ctx, backlinks)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	result := make([]app.NoteLink, len(backlinkNotes))
	for i, linkedNote := range backlinkNotes {
		var icon string
		if linkedNote.Icon != nil {
			icon = *linkedNote.Icon
		}
		result[i] = app.NoteLink{
			ID:   linkedNote.ID,
			Name: linkedNote.Name,
			Icon: icon,
		}
	}
	return result, nil
}
