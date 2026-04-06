package pg

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

type NoteRepo struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	publisher     *Publisher
	inTransaction bool
}

var _ domain.NoteRepo = (*NoteRepo)(nil)

func NewNoteRepo(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	publisher *Publisher,
	inTransaction bool,
) *NoteRepo {
	return &NoteRepo{
		pgxPool:       pgxPool,
		queries:       queries,
		publisher:     publisher,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionNoteRepo(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
) *NoteRepo {
	return NewNoteRepo(pgxPool, queries, nil, false)
}

var ProvideNoteRepo = NewNoTransactionNoteRepo

func (n *NoteRepo) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Note, error) {
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

func (n *NoteRepo) GetMany(ctx context.Context, params *domain.NoteRepoGetManyParams) ([]*domain.Note, error) {
	var ids *[]uuid.UUID
	if len(params.IDs()) > 0 {
		paramIDs := params.IDs()
		ids = &paramIDs
	}

	var workspaceID *uuid.UUID
	if params.WorkspaceID() != nil {
		workspaceID = params.WorkspaceID()
	}

	var trashedBy *string
	if params.TrashedBy() != nil {
		trashedByStr := params.TrashedBy().String()
		trashedBy = &trashedByStr
	}

	isNotTrashed := true // Default: filter for non-trashed notes
	if params.IsTrashed() != nil {
		isNotTrashed = !*params.IsTrashed() // If IsTrashed=true, then IsNotTrashed=false
	}

	notes, err := n.queries.GetNotesByParams(ctx, &pgsqlc.GetNotesByParamsParams{
		IDs:          ids,
		WorkspaceID:  workspaceID,
		TrashedBy:    trashedBy,
		IsNotTrashed: isNotTrashed,
		ForUpdate:    params.ForUpdate(),
	})
	if err != nil {
		return nil, toErr(err)
	}

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
	var trashed *domain.Trashed
	if note.TrashedBy != nil && note.TrashedAt != nil {
		trashed = domain.NewTrashed(
			domain.TrashedBy(*note.TrashedBy),
			*note.TrashedAt,
		)
	}
	var tags []string
	if note.Tags != nil {
		tags = note.Tags
	}
	return domain.UnmarshalNote(
		note.ID,
		note.Name,
		note.Icon,
		tags,
		uint64(note.Size),
		note.FolderID,
		links,
		trashed,
	)
}

// TODO: It doesn't save the outgoing links
func (n *NoteRepo) Save(ctx context.Context, note *domain.Note) (cerr error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       n.pgxPool,
		queries:       n.queries,
		inTransaction: n.inTransaction,
	}, func(params *RunInTxFnparams) error {
		err := params.queries.SaveNote(ctx, &pgsqlc.SaveNoteParams{
			ID:        note.ID(),
			Name:      note.Name(),
			Icon:      note.Icon(),
			FolderID:  note.FolderID(),
			Tags:      note.Tags(),
			Size:      int32(note.Size()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			TrashedBy: note.TrashedByString(),
			TrashedAt: note.TrashedAt(),
		})
		if err != nil {
			return toErr(err)
		}
		if err := params.queries.CreateTempTableNoteLinks(ctx); err != nil {
			return toErr(err)
		}
		saveNoteLinkParams := make([]*pgsqlc.InsertTempNoteLinksParams, len(note.OutgoingLinks()))
		for i, targetID := range note.OutgoingLinks() {
			saveNoteLinkParams[i] = &pgsqlc.InsertTempNoteLinksParams{
				SourceID: note.ID(),
				TargetID: targetID,
			}
		}
		affected, err := params.queries.InsertTempNoteLinks(ctx, saveNoteLinkParams)
		if err != nil {
			return toErr(err)
		}
		if affected != int64(len(note.OutgoingLinks())) {
			return fmt.Errorf("not all note links were inserted into temp table")
		}
		if err := params.queries.DeleteObsoleteNoteLinks(ctx); err != nil {
			return toErr(err)
		}
		if err := params.queries.SaveFromTempNoteLinks(ctx); err != nil {
			return toErr(err)
		}
		if err := params.publisher.Publish(ctx, note.PopEvents()...); err != nil {
			return fmt.Errorf("failed to publish events: %w", err)
		}
		return nil
	})
}

// TODO: It doesn't save the outgoing links
func (n *NoteRepo) SaveMany(ctx context.Context, notes []*domain.Note) (cerr error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       n.pgxPool,
		queries:       n.queries,
		inTransaction: n.inTransaction,
	}, func(params *RunInTxFnparams) error {
		if err := params.queries.CreateTempTableNotes(ctx); err != nil {
			return fmt.Errorf("failed to create temp table for notes: %w", toErr(err))
		}
		saveNoteParams := make([]*pgsqlc.InsertTempNotesParams, len(notes))
		for i, note := range notes {
			saveNoteParams[i] = &pgsqlc.InsertTempNotesParams{
				ID:        note.ID(),
				Name:      note.Name(),
				Icon:      note.Icon(),
				FolderID:  note.FolderID(),
				Tags:      note.Tags(),
				Size:      int32(note.Size()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				TrashedBy: note.TrashedByString(),
				TrashedAt: note.TrashedAt(),
			}
		}
		affected, err := params.queries.InsertTempNotes(ctx, saveNoteParams)
		if err != nil {
			return fmt.Errorf("failed to insert notes into temp table: %w", toErr(err))
		}
		if affected != int64(len(notes)) {
			return fmt.Errorf("not all notes were inserted into temp table (expected %d, got %d)", len(notes), affected)
		}
		if err = params.queries.SaveFromTempNotes(ctx); err != nil {
			return fmt.Errorf("failed to save notes from temp table: %w", toErr(err))
		}
		events := make([]domain.Event, 0)
		for _, note := range notes {
			events = append(events, note.PopEvents()...)
		}
		if err := params.publisher.Publish(ctx, events...); err != nil {
			return fmt.Errorf("failed to publish events: %w", err)
		}
		return nil
	})
}

func (n *NoteRepo) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, error) {
	count, err := n.queries.CountNotesInWorkspaceByIDs(ctx, &pgsqlc.CountNotesInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if notes are in workspace: %w", toErr(err))
	}
	return count == int64(len(ids)), nil
}

func (n *NoteRepo) PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error {
	err := n.queries.PermanentlyDeleteNoteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete note: %w", toErr(err))
	}
	return nil
}

func (n *NoteRepo) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error {
	err := n.queries.PermanentlyDeleteNotesByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("failed to permanently delete notes: %w", toErr(err))
	}
	return nil
}

func (n *NoteRepo) GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	workspaceID, err := n.queries.GetWorkspaceIDByNoteID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get workspace id for note: %w", toErr(err))
	}
	return workspaceID, nil
}
