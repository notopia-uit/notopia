package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type Note struct {
	queries *pgsqlc.Queries
}

var _ domain.NoteRepo = (*Note)(nil)

func NewNote(queries *pgsqlc.Queries) *Note {
	return &Note{
		queries: queries,
	}
}

var ProvideNote = NewNote

func (n *Note) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	noteResult, err := n.queries.GetNote(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNoteNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	outgoingLinksResult, err := n.queries.GetNoteOutgoingLinks(ctx, id)
	if err != nil {
		return nil, toDomainError(err)
	}
	return noteToDomain(noteResult, outgoingLinksResult), nil
}

func (n *Note) Save(ctx context.Context, note *domain.Note) error {
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

func (n *Note) GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]domain.Note, error) {
	noteResults, err := n.queries.GetTrashedNotesByWorkspaceID(ctx, workspaceID)
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

func (n *Note) PermanentlyDeleteByID(ctx context.Context, id uuid.UUID) error {
	err := n.queries.PermanentlyDeleteNoteByID(ctx, id)
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (n *Note) PermanentlyDeleteByIDs(ctx context.Context, ids uuid.UUIDs) error {
	err := n.queries.PermanentlyDeleteNotesByIDs(ctx, ids)
	if err != nil {
		return toDomainError(err)
	}
	return nil
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
		uint(note.Size),
		note.FolderID,
		outgoingLinks,
		trashed,
	)
}
