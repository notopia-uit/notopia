package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type UnpublishNote struct {
	ID uuid.UUID
}

type UnpublishNoteHandler struct {
	noteRepo domain.NoteRepo
}

func NewUnpublishNoteHandler(noteRepo domain.NoteRepo) *UnpublishNoteHandler {
	return &UnpublishNoteHandler{noteRepo: noteRepo}
}

var ProvideUnpublishNoteHandler = NewUnpublishNoteHandler

func (h *UnpublishNoteHandler) Handle(ctx context.Context, cmd *UnpublishNote) error {
	// WARN: Handler is incomplete - domain.Note has no Unpublish() method.
	// TODO: domain.Note has no Unpublish() method. Add Unpublish() to domain.Note and a
	// published field, then call note.Unpublish() here before Save.
	// This mirrors PublishNote handler - requires same domain.Note.published field addition.
	// Steps:
	// 1. Add `published bool` field to domain.Note struct (done with Publish handler)
	// 2. Add Unpublish() method to Note: func (n *Note) Unpublish() { n.published = false }
	// 3. Update Note.Unmarshal() to accept published parameter (done with Publish handler)
	// 4. Update persistence layer (done with Publish handler)
	// 5. Implement this handler to call note.Unpublish(), add event, and save
	// TODO: note.Unpublish() not yet implemented
	return nil
}
