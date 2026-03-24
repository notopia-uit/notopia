package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
}

var _ domain.NoteRepo = (*Note)(nil)

func NewNote(queries *pgsqlc.Queries) *Note {
	return &Note{queries: queries}
}

var ProvideNote = NewNote

func (n *Note) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*domain.Note, errs.Error) {
	var noteResult *pgsqlc.Note
	var err error
	if forUpdate {
		noteResult, err = n.queries.GetNoteForUpdate(ctx, id)
	} else {
		noteResult, err = n.queries.GetNote(ctx, id)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	outgoingLinksResult, err := n.queries.GetNoteOutgoingLinks(ctx, id)
	if err != nil {
		return nil, toDomainError(err)
	}
	return noteToDomain(noteResult, outgoingLinksResult), nil
}

func (n *Note) GetByIDs(ctx context.Context, ids uuid.UUIDs, forUpdate bool) ([]domain.Note, errs.Error) {
	var noteResults []*pgsqlc.Note
	var err error
	if forUpdate {
		noteResults, err = n.queries.GetNotesForUpdate(ctx, ids)
	} else {
		noteResults, err = n.queries.GetNotes(ctx, ids)
	}
	if err != nil {
		return nil, toDomainError(err)
	}
	outgoingLinkResults, err := n.queries.GetNotesOutgoingLinks(ctx, ids)
	if err != nil {
		return nil, toDomainError(err)
	}
	outgoingLinksMap := make(map[uuid.UUID]uuid.UUIDs)
	for _, outgoingLink := range outgoingLinkResults {
		outgoingLinksMap[outgoingLink.SourceID] = append(outgoingLinksMap[outgoingLink.SourceID], outgoingLink.TargetID)
	}
	notes := make([]domain.Note, len(noteResults))
	for i, note := range noteResults {
		notes[i] = *noteToDomain(note, outgoingLinksMap[note.ID])
	}
	return notes, nil
}

func (n *Note) GetByWorkspaceID(ctx context.Context, params domain.NoteRepoGetByWorkspaceIDParams) ([]domain.Note, errs.Error) {
	var trashedBy *string
	if params.TrashedBy != nil {
		trashedBy = new(params.TrashedBy.String())
	}

	noteResults, err := n.queries.GetNotesInWorkspace(ctx, &pgsqlc.GetNotesInWorkspaceParams{
		WorkspaceID: params.WorkspaceID,
		TrashedBy:   trashedBy,
	})
	if err != nil {
		return nil, toDomainError(err)
	}
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
	notes := make([]domain.Note, len(noteResults))
	for i, note := range noteResults {
		notes[i] = *noteToDomain(note, outgoingLinksMap[note.ID])
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

func (n *Note) SaveMany(ctx context.Context, notes []domain.Note) errs.Error {
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

func (n *Note) GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]domain.Note, errs.Error) {
	return n.GetByWorkspaceID(ctx, domain.NoteRepoGetByWorkspaceIDParams{
		WorkspaceID: workspaceID,
		TrashedBy:   &domain.TrashedByPurpose,
	})
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
