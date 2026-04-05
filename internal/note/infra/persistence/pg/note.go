package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	inTransaction bool
}

var _ domain.NoteRepo = (*Note)(nil)

func NewNote(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	inTransaction bool,
) *Note {
	return &Note{
		pgxPool:       pgxPool,
		queries:       queries,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionNote(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
) *Note {
	return NewNote(pgxPool, queries, false)
}

var ProvideNote = NewNoTransactionNote

func (n *Note) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Note, errs.Error) {
	note, err := n.queries.GetNoteByID(ctx, pgsqlc.GetNoteByIDParams{
		ID:        id,
		ForUpdate: forUpdate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(id, err)
		}
		return nil, toDomainError(err)
	}

	// Fetch outgoing links separately
	links, err := n.queries.GetNoteOutgoingLinks(ctx, &pgsqlc.GetNoteOutgoingLinksParams{
		SourceID:  &id,
		SourceIDs: nil,
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	return noteToDomain(note, links), nil
}

func (n *Note) GetMany(ctx context.Context, params *domain.NoteRepoGetManyParams) ([]*domain.Note, errs.Error) {
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
		return nil, toDomainError(err)
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
			return nil, toDomainError(err)
		}
		result[i] = noteToDomain(note, links)
	}
	return result, nil
}

func noteToDomain(note *pgsqlc.Note, links []uuid.UUID) *domain.Note {
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
func (n *Note) Save(ctx context.Context, note *domain.Note) (cerr errs.Error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       n.pgxPool,
		queries:       n.queries,
		inTransaction: n.inTransaction,
	}, func(queries *pgsqlc.Queries) errs.Error {
		err := queries.SaveNote(ctx, &pgsqlc.SaveNoteParams{
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
			return toDomainError(err)
		}
		if err := queries.CreateTempTableNoteLinks(ctx); err != nil {
			return toDomainError(err)
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
			return toDomainError(err)
		}
		if affected != int64(len(note.OutgoingLinks())) {
			return toDomainError(errors.New("not all note links were inserted into temp table"))
		}
		if err := queries.DeleteObsoleteNoteLinks(ctx); err != nil {
			return toDomainError(err)
		}
		if err := queries.SaveFromTempNoteLinks(ctx); err != nil {
			return toDomainError(err)
		}
		return nil
	})
}

// TODO: It doesn't save the outgoing links
func (n *Note) SaveMany(ctx context.Context, notes []*domain.Note) (cerr errs.Error) {
	return runInTx(ctx, &runInTxParams{
		pgxPool:       n.pgxPool,
		queries:       n.queries,
		inTransaction: n.inTransaction,
	}, func(queries *pgsqlc.Queries) errs.Error {
		if err := queries.CreateTempTableNotes(ctx); err != nil {
			return toDomainError(err)
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
		affected, err := queries.InsertTempNotes(ctx, saveNoteParams)
		if err != nil {
			return toDomainError(err)
		}
		if affected != int64(len(notes)) {
			return toDomainError(errors.New("not all notes were inserted into temp table"))
		}
		if err = queries.SaveFromTempNotes(ctx); err != nil {
			return toDomainError(err)
		}
		return nil
	})
}

func (n *Note) AreAllInWorkspace(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) (bool, errs.Error) {
	count, err := n.queries.CountNotesInWorkspaceByIDs(ctx, &pgsqlc.CountNotesInWorkspaceByIDsParams{
		IDs:         ids,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, toDomainError(err)
	}
	return count == int64(len(ids)), nil
}

func (n *Note) PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) errs.Error {
	err := n.queries.PermanentlyDeleteNoteByID(ctx, id)
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (n *Note) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) errs.Error {
	err := n.queries.PermanentlyDeleteNotesByIDs(ctx, ids)
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (n *Note) GetWorkspaceIDByID(ctx context.Context, id uuid.UUID) (uuid.UUID, errs.Error) {
	workspaceID, err := n.queries.GetWorkspaceIDByNoteID(ctx, id)
	if err != nil {
		return uuid.Nil, toDomainError(err)
	}
	return workspaceID, nil
}
