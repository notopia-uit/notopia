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

func NewNoteRepo(queries *pgsqlc.Queries) domain.NoteRepo {
	return &Note{
		queries: queries,
	}
}

func (n *Note) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	result, err := n.queries.GetNote(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNoteNotFound(id, err)
		}
		return nil, toDomainError(err)
	}
	return noteToDomain(result), nil
}

func (n *Note) Save(ctx context.Context, note *domain.Note) error {
	var trashedBy *string
	if note.TrashedBy() != nil {
		trashedByStr := string(*note.TrashedBy())
		trashedBy = &trashedByStr
	}

	err := n.queries.SaveNote(ctx, &pgsqlc.SaveNoteParams{
		ID:        note.ID(),
		Name:      note.Name(),
		Icon:      note.Icon(),
		FolderID:  note.FolderID(),
		Tags:      note.Tags(),
		Size:      int32(note.Size()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedBy: trashedBy,
		DeletedAt: note.TrashedAt(),
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (n *Note) GetTrashedByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]domain.Note, error) {
	results, err := n.queries.GetTrashedNotesByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, toDomainError(err)
	}
	notes := make([]domain.Note, len(results))
	for i, note := range results {
		notes[i] = *noteToDomain(note)
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

func noteToDomain(note *pgsqlc.Note) *domain.Note {
	var trashedBy *domain.TrashedBy
	if note.DeletedBy != nil {
		trashedByVal := domain.TrashedBy(*note.DeletedBy)
		trashedBy = &trashedByVal
	}

	return domain.UnmarshalNote(
		note.ID,
		note.Name,
		note.Icon,
		note.Tags,
		uint(note.Size),
		note.FolderID,
		uuid.UUIDs([]uuid.UUID{}),
		trashedBy,
		note.DeletedAt,
	)
}
