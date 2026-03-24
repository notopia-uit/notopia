package pg

import (
	"context"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgjet/public/table"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
	db      qrm.DB
}

var _ domain.NoteRepo = (*Note)(nil)

func NewNote(queries *pgsqlc.Queries, db qrm.DB) *Note {
	return &Note{
		queries: queries,
		db:      db,
	}
}

var ProvideNote = NewNote

func (n *Note) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Note, errs.Error) {
	stmt := SELECT(table.Notes.AllColumns).
		FROM(table.Notes).
		WHERE(table.Notes.ID.EQ(UUID(id)))
	if forUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*pgsqlc.Note
	err := stmt.QueryContext(ctx, n.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(dest) == 0 {
		return nil, errs.NewNoteNotFound(id, pgx.ErrNoRows)
	}
	noteResult := dest[0]

	outgoingLinksResult, err := n.queries.GetNoteOutgoingLinks(ctx, id)
	if err != nil {
		return nil, toDomainError(err)
	}
	return noteToDomain(noteResult, outgoingLinksResult), nil
}

func (n *Note) GetMany(ctx context.Context, params domain.NoteRepoGetManyParams) ([]*domain.Note, errs.Error) {
	condition := Bool(true)
	if params.WorkspaceID != nil {
		condition = condition.AND(
			table.Notes.FolderID.IN(
				SELECT(table.Folders.ID).
					FROM(table.Folders).
					WHERE(table.Folders.WorkspaceID.EQ(UUID(params.WorkspaceID))),
			),
		)
	}
	if len(params.IDs) > 0 {
		var idExprs []Expression
		for _, id := range params.IDs {
			idExprs = append(idExprs, UUID(id))
		}
		condition = condition.AND(table.Notes.ID.IN(idExprs...))
	}
	if params.TrashedBy != nil {
		condition = condition.AND(table.Notes.TrashedBy.EQ(String(params.TrashedBy.String())))
	}

	stmt := SELECT(table.Notes.AllColumns).
		FROM(table.Notes).
		WHERE(condition)
	if params.ForUpdate {
		stmt = stmt.FOR(UPDATE())
	}

	var dest []*pgsqlc.Note
	err := stmt.QueryContext(ctx, n.db, &dest)
	if err != nil {
		return nil, toDomainError(err)
	}

	if len(dest) == 0 {
		return []*domain.Note{}, nil
	}

	noteResults := dest

	noteIDs := make([]uuid.UUID, len(noteResults))
	for i, note := range noteResults {
		noteIDs[i] = note.ID
	}

	outgoingLinkResults, err := n.queries.GetNotesOutgoingLinks(ctx, noteIDs)
	if err != nil {
		return nil, toDomainError(err)
	}
	outgoingLinksMap := make(map[uuid.UUID]uuid.UUIDs)
	for _, outgoingLink := range outgoingLinkResults {
		outgoingLinksMap[outgoingLink.SourceID] = append(outgoingLinksMap[outgoingLink.SourceID], outgoingLink.TargetID)
	}
	notes := make([]*domain.Note, len(noteResults))
	for i, note := range noteResults {
		notes[i] = noteToDomain(note, outgoingLinksMap[note.ID])
	}
	return notes, nil
}

func (n *Note) Save(ctx context.Context, note *domain.Note) errs.Error {
	err := n.queries.SaveNote(ctx, &pgsqlc.SaveNoteParams{
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
	return nil
}

func (n *Note) SaveMany(ctx context.Context, notes []*domain.Note) errs.Error {
	err := n.queries.CreateTempTableNotes(ctx)
	if err != nil {
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
	affected, err := n.queries.InsertTempNotes(ctx, saveNoteParams)
	if err != nil {
		return toDomainError(err)
	}
	if affected != int64(len(notes)) {
		return toDomainError(errors.New("not all notes were inserted into temp table"))
	}
	err = n.queries.SaveFromTempNotes(ctx)
	if err != nil {
		return toDomainError(err)
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

func noteToDomain(note *pgsqlc.Note, outgoingLinks uuid.UUIDs) *domain.Note {
	var trashed *domain.Trashed
	if note.TrashedBy != nil && note.TrashedAt != nil {
		trashed = domain.NewTrashed(
			domain.TrashedBy(*note.TrashedBy),
			*note.TrashedAt,
		)
	}
	return domain.UnmarshalNote(
		note.ID,
		note.Name,
		note.Icon,
		note.Tags,
		uint64(note.Size),
		note.FolderID,
		outgoingLinks,
		trashed,
	)
}
