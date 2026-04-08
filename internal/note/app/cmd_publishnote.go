package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishNote struct {
	ID uuid.UUID
}

type PublishNoteHandler struct {
	noteRepo domain.NoteRepo
}

func NewPublishNoteHandler(noteRepo domain.NoteRepo) *PublishNoteHandler {
	return &PublishNoteHandler{noteRepo: noteRepo}
}

var ProvidePublishNoteHandler = NewPublishNoteHandler

func (h *PublishNoteHandler) Handle(ctx context.Context, cmd *PublishNote) error {
	// WARN: Handler is incomplete - domain.Note has no Publish() method.
	// TODO: domain.Note has no Publish() method. Add Publish() to domain.Note and a
	// published field, then call note.Publish() here before Save.
	// Steps:
	// 1. Add `published bool` field to domain.Note struct
	// 2. Add Publish() method to Note: func (n *Note) Publish() { n.published = true }
	// 3. Update Note.Unmarshal() to accept published parameter
	// 4. Update persistence layer to store/retrieve published field
	// 5. Implement this handler to call note.Publish(), add event, and save
	// TODO: note.Publish() not yet implemented
	return nil
}
