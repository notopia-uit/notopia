package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/model"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/table"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	pgxPool       *pgxpool.Pool
	queries       *pgsqlc.Queries
	db            qrm.DB
	inTransaction bool
}

var _ domain.NoteRepo = (*Note)(nil)

func NewNote(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	db qrm.DB,
	inTransaction bool,
) *Note {
	return &Note{
		pgxPool:       pgxPool,
		queries:       queries,
		db:            db,
		inTransaction: inTransaction,
	}
}

func NewNoTransactionNote(
	pgxPool *pgxpool.Pool,
	queries *pgsqlc.Queries,
	db qrm.DB,
) *Note {
	return NewNote(pgxPool, queries, db, false)
}

var ProvideNote = NewNoTransactionNote

type GetNoteResult struct {
	model.Notes
	OutgoingLinks uuid.UUIDs `alias:"note_links.target_id"`
}

func (n *Note) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Note, errs.Error) {
	stmt := SELECT(table.Notes.AllColumns).
		FROM(
			table.Notes.
				LEFT_JOIN(table.NoteLinks, table.NoteLinks.SourceID.EQ(table.Notes.ID)),
		).
		WHERE(table.Notes.ID.EQ(UUID(id)))
	if forUpdate {
		stmt = stmt.FOR(UPDATE())
	}
	var dest *GetNoteResult
	err := stmt.QueryContext(ctx, n.db, dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, errs.NewNoteNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	return noteToDomain(dest), nil
}

func (n *Note) GetMany(ctx context.Context, params *domain.NoteRepoGetManyParams) ([]*domain.Note, errs.Error) {
	condition := Bool(true)
	if params.WorkspaceID() != nil {
		condition = condition.AND(
			table.Notes.FolderID.IN(
				SELECT(table.Folders.ID).
					FROM(table.Folders).
					WHERE(table.Folders.WorkspaceID.EQ(UUID(*params.WorkspaceID()))),
			),
		)
	}
	if len(params.IDs()) > 0 {
		var idExprs []Expression
		for _, id := range params.IDs() {
			idExprs = append(idExprs, UUID(id))
		}
		condition = condition.AND(table.Notes.ID.IN(idExprs...))
	}
	if params.TrashedBy() != nil {
		condition = condition.AND(table.Notes.TrashedBy.EQ(String(params.TrashedBy().String())))
	}
	if params.IsTrashed() != nil {
		condition = condition.AND(table.Notes.TrashedAt.IS_NULL())
	}

	stmt := SELECT(table.Notes.AllColumns).
		FROM(
			table.Notes.
				LEFT_JOIN(table.NoteLinks, table.NoteLinks.SourceID.EQ(table.Notes.ID)),
		).
		WHERE(condition)
	if params.ForUpdate() {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*GetNoteResult
	err := stmt.QueryContext(ctx, n.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(dest) == 0 {
		return []*domain.Note{}, nil
	}

	notes := make([]*domain.Note, len(dest))
	for i, noteResult := range dest {
		notes[i] = noteToDomain(noteResult)
	}
	return notes, nil
}

func noteToDomain(note *GetNoteResult) *domain.Note {
	var trashed *domain.Trashed
	if note.TrashedBy != nil && note.TrashedAt != nil {
		trashed = domain.NewTrashed(
			domain.TrashedBy(*note.TrashedBy),
			*note.TrashedAt,
		)
	}
	var tags []string
	if note.Tags != nil {
		tags = *note.Tags
	}
	return domain.UnmarshalNote(
		note.ID,
		note.Name,
		note.Icon,
		tags,
		uint64(note.Size),
		note.FolderID,
		note.OutgoingLinks,
		trashed,
	)
}

// TODO: It doesn't save the outgoing links
func (n *Note) Save(ctx context.Context, note *domain.Note) (cerr errs.Error) {
	var queries *pgsqlc.Queries
	var tx pgx.Tx
	var err error
	if !n.inTransaction {
		tx, err = n.pgxPool.Begin(ctx)
		if err != nil {
			return toDomainError(err)
		}
		queries = n.queries.WithTx(tx)
		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				cerr = errs.NewPersistenceInternal("failed to rollback transaction", fmt.Errorf("%w: %v", cerr, err))
			}
		}()
	} else {
		queries = n.queries
	}
	err = queries.SaveNote(ctx, &pgsqlc.SaveNoteParams{
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
	if !n.inTransaction {
		if err := tx.Commit(ctx); err != nil {
			return errs.NewPersistenceInternal("failed to commit transaction", err)
		}
	}
	return nil
}

// TODO: It doesn't save the outgoing links
func (n *Note) SaveMany(ctx context.Context, notes []*domain.Note) (cerr errs.Error) {
	var queries *pgsqlc.Queries
	var tx pgx.Tx
	var err error
	if !n.inTransaction {
		tx, err = n.pgxPool.Begin(ctx)
		if err != nil {
			return toDomainError(err)
		}
		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				cerr = errs.NewPersistenceInternal("failed to rollback transaction", fmt.Errorf("%w: %v", cerr, err))
			}
		}()
		queries = n.queries.WithTx(tx)
	} else {
		queries = n.queries
	}
	if err = queries.CreateTempTableNotes(ctx); err != nil {
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
	if !n.inTransaction {
		if err := tx.Commit(ctx); err != nil {
			return errs.NewPersistenceInternal("failed to commit transaction", err)
		}
	}
	return nil
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
