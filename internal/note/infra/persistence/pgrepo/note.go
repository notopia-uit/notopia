package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

// TODO: what, it has so many fmt error? we have to map to toErr only
type Note struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	publisher     Publisher
	inTransaction bool
}

var _ domain.NoteRepo = (*Note)(nil)

func NewNote(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	publisher Publisher,
	inTransaction bool,
) *Note {
	return &Note{
		pgxPool:       pgxPool,
		queries:       queries,
		publisher:     publisher,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionNote(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
) *Note {
	return NewNote(pgxPool, queries, nil, false)
}

var ProvideNote = NewNoTransactionNote

func (n *Note) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Note, error) {
	note, err := n.queries.GetNoteByID(ctx, pgsqlc.GetNoteByIDParams{
		ID:        id,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(id, err)
		}
		return nil, toErr(err)
	}

	// Fetch outgoing links separately
	links, err := n.queries.GetNoteOutgoingLinks(ctx, &pgsqlc.GetNoteOutgoingLinksParams{
		SourceID:  &id,
		SourceIDs: nil,
	})
	if err != nil {
		return nil, toErr(err)
	}

	return noteToDomainRepo(note, links), nil
}

func (n *Note) GetMany(ctx context.Context, params *domain.NoteRepoGetManyParams) ([]*domain.Note, error) {
	var ids *[]uuid.UUID
	if len(params.IDs) > 0 {
		ids = &params.IDs
	}

	var workspaceID *uuid.UUID
	if params.WorkspaceID != uuid.Nil {
		workspaceID = &params.WorkspaceID
	}

	var trashedBy *string
	if params.TrashedBy != domain.TrashedByUnspecified {
		var ok bool
		trashedBy, ok = fromDomainTrashedBy(params.TrashedBy)
		if !ok {
			return nil, fmt.Errorf("invalid trashed by value: %v", params.TrashedBy)
		}
	}

	notes, err := n.queries.GetNotesByParams(ctx, &pgsqlc.GetNotesByParamsParams{
		IDs:            ids,
		WorkspaceID:    workspaceID,
		TrashedBy:      trashedBy,
		OnlyNonTrashed: params.NotTrashedOnly,
		OnlyTrashed:    params.TrashOnly,
		ForUpdate:      params.ForUpdate,
	})
	if err != nil {
		return nil, toErr(err)
	}

	// FIXME: Hey, this is N+1 query
	// Query links per note (type-safe approach)
	result := make([]*domain.Note, len(notes))
	for i, note := range notes {
		// Query links for this specific note
		links, err := n.queries.GetNoteOutgoingLinks(ctx, &pgsqlc.GetNoteOutgoingLinksParams{
			SourceID:  &note.ID,
			SourceIDs: nil,
		})
		if err != nil {
			return nil, toErr(err)
		}
		result[i] = noteToDomainRepo(note, links)
	}
	return result, nil
}

func noteToDomainRepo(note *pgsqlc.Note, links []uuid.UUID) *domain.Note {
	var icon string
	if note.Icon != nil {
		icon = *note.Icon
	}
	var trashed *domain.Trashed
	if note.TrashedBy != nil && note.TrashedAt != nil {
		trashed = domain.NewTrashed(
			toDomainTrashedBy(*note.TrashedBy),
			*note.TrashedAt,
		)
	}
	return domain.UnmarshalNote(
		note.ID,
		note.Name,
		icon,
		note.Tags,
		uint64(note.Size),
		note.FolderID,
		links,
		trashed,
		false,
	)
}

// TODO: It doesn't save the outgoing links
func (n *Note) Save(ctx context.Context, note *domain.Note) error {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       n.pgxPool,
		queries:       n.queries,
		publisher:     n.publisher,
		inTransaction: n.inTransaction,
	}, func(params *RunInTxFnparams) error {
		queries := params.queries
		if note.Deleted() {
			if err := queries.PermanentlyDeleteNoteByID(ctx, note.ID()); err != nil {
				return toErr(err)
			}
		} else {
			var icon *string
			if note.Icon() != "" {
				icon = new(note.Icon())
			}
			var trashedBy *string
			var trashedAt *time.Time
			if note.IsTrashed() {
				by := note.TrashedBy().String()
				trashedBy = &by
				t := note.TrashedAt()
				trashedAt = &t
			}
			err := queries.SaveNote(ctx, &pgsqlc.SaveNoteParams{
				ID:        note.ID(),
				Name:      note.Name(),
				Icon:      icon,
				FolderID:  note.FolderID(),
				Tags:      note.Tags(),
				Size:      int32(note.Size()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				TrashedBy: trashedBy,
				TrashedAt: trashedAt,
			})
			if err != nil {
				return toErr(err)
			}
			if err := queries.CreateTempTableNoteLinks(ctx); err != nil {
				return toErr(err)
			}
			saveNoteLinkParams := make([]*pgsqlc.InsertTempNoteLinksParams, len(note.OutgoingLinks()))
			for i, targetID := range note.OutgoingLinks() {
				saveNoteLinkParams[i] = &pgsqlc.InsertTempNoteLinksParams{
					SourceID: note.ID(),
					TargetID: targetID,
				}
			}
			affected, err := queries.InsertTempNoteLinks(ctx, saveNoteLinkParams)
			if err != nil {
				return toErr(err)
			}
			if affected != int64(len(note.OutgoingLinks())) {
				return fmt.Errorf("not all note links were inserted into temp table")
			}
			if err := queries.DeleteObsoleteNoteLinks(ctx); err != nil {
				return toErr(err)
			}
			if err := queries.SaveFromTempNoteLinks(ctx); err != nil {
				return toErr(err)
			}
		}
		workspaceID, err := queries.GetWorkspaceIDByNoteID(ctx, note.ID())
		if err != nil {
			return toErr(err)
		}
		for _, event := range note.PopEvents() {
			if err := params.publisher.PublishWorkspaceItem(
				ctx,
				event,
				PublishWorkspaceItemParams{
					WorkspaceID: workspaceID,
					AggregateID: note.ID(),
				},
			); err != nil {
				return fmt.Errorf("failed to publish events: %w", err)
			}
		}
		return nil
	})
}

// TODO: It doesn't save the outgoing links
func (n *Note) SaveMany(ctx context.Context, notes []*domain.Note) error {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       n.pgxPool,
		queries:       n.queries,
		publisher:     n.publisher,
		inTransaction: n.inTransaction,
	}, func(params *RunInTxFnparams) error {
		var deleteIDs []uuid.UUID
		var upsertNotes []*domain.Note
		var allOutgoingLinks []*pgsqlc.InsertTempNoteLinksParams

		for _, note := range notes {
			if note.Deleted() {
				deleteIDs = append(deleteIDs, note.ID())
			} else {
				upsertNotes = append(upsertNotes, note)
				for _, targetID := range note.OutgoingLinks() {
					allOutgoingLinks = append(allOutgoingLinks, &pgsqlc.InsertTempNoteLinksParams{
						SourceID: note.ID(),
						TargetID: targetID,
					})
				}
			}
		}

		if err := n.deleteMany(params.queries, ctx, deleteIDs); err != nil {
			return err
		}

		if err := n.upsertMany(params.queries, ctx, upsertNotes, allOutgoingLinks); err != nil {
			return err
		}

		noteIDs := make([]uuid.UUID, len(notes))
		for i, note := range notes {
			noteIDs[i] = note.ID()
		}
		noteIDworkspaceIDPairs, err := params.queries.GetWorkspaceIDsByNoteIDs(ctx, noteIDs)
		noteIDWorkspaceIDMap := make(map[uuid.UUID]uuid.UUID, len(noteIDworkspaceIDPairs))
		for _, pair := range noteIDworkspaceIDPairs {
			noteIDWorkspaceIDMap[pair.ID] = pair.WorkspaceID
		}
		if err != nil {
			return toErr(err)
		}
		for _, note := range notes {
			workspaceID, ok := noteIDWorkspaceIDMap[note.ID()]
			if !ok {
				return fmt.Errorf("failed to find workspace id for note id %s", note.ID())
			}
			for _, event := range note.PopEvents() {
				if err := params.publisher.PublishWorkspaceItem(
					ctx,
					event,
					PublishWorkspaceItemParams{
						WorkspaceID: workspaceID,
						AggregateID: note.ID(),
					},
				); err != nil {
					return fmt.Errorf("failed to publish events: %w", err)
				}
			}
		}
		return nil
	})
}

func (n *Note) deleteMany(queries *pgsqlc.Queries, ctx context.Context, deleteIDs []uuid.UUID) error {
	if len(deleteIDs) > 0 {
		if err := queries.PermanentlyDeleteNotesByIDs(ctx, deleteIDs); err != nil {
			return fmt.Errorf("failed bulk delete: %w", toErr(err))
		}
	}
	return nil
}

func (n *Note) upsertMany(queries *pgsqlc.Queries, ctx context.Context, upsertNotes []*domain.Note, allOutgoingLinks []*pgsqlc.InsertTempNoteLinksParams) error {
	if len(upsertNotes) > 0 {
		if err := queries.CreateTempTableNotes(ctx); err != nil {
			return toErr(err)
		}

		saveNoteParams := make([]*pgsqlc.InsertTempNotesParams, len(upsertNotes))
		for i, note := range upsertNotes {
			var icon *string
			if note.Icon() != "" {
				icon = new(note.Icon())
			}
			var trashedBy *string
			var trashedAt *time.Time
			if note.IsTrashed() {
				trashedBy = new(note.TrashedBy().String())
				trashedAt = new(note.TrashedAt())
			}
			saveNoteParams[i] = &pgsqlc.InsertTempNotesParams{
				ID:        note.ID(),
				Name:      note.Name(),
				Icon:      icon,
				FolderID:  note.FolderID(),
				Tags:      note.Tags(),
				Size:      int32(note.Size()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				TrashedBy: trashedBy,
				TrashedAt: trashedAt,
			}
		}

		if _, err := queries.InsertTempNotes(ctx, saveNoteParams); err != nil {
			return toErr(err)
		}
		if err := queries.SaveFromTempNotes(ctx); err != nil {
			return toErr(err)
		}

		if len(allOutgoingLinks) > 0 {
			if err := queries.CreateTempTableNoteLinks(ctx); err != nil {
				return toErr(err)
			}
			if _, err := queries.InsertTempNoteLinks(ctx, allOutgoingLinks); err != nil {
				return toErr(err)
			}
			if err := queries.DeleteObsoleteNoteLinks(ctx); err != nil {
				return toErr(err)
			}
			if err := queries.SaveFromTempNoteLinks(ctx); err != nil {
				return toErr(err)
			}
		}
	}
	return nil
}

func (n *Note) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error) {
	count, err := n.queries.CountNotesInWorkspaceByIDs(ctx, &pgsqlc.CountNotesInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if notes are in workspace: %w", toErr(err))
	}
	return count == int64(len(ids)), nil
}

func (n *Note) GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	workspaceID, err := n.queries.GetWorkspaceIDByNoteID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get workspace id for note: %w", toErr(err))
	}
	return workspaceID, nil
}
