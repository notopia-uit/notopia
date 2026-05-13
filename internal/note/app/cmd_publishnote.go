package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishNote struct {
	ID     uuid.UUID
	UserID string
}

type PublishNoteHandler struct {
	noteRepo domain.NoteRepo
}

func NewPublishNoteHandler(noteRepo domain.NoteRepo) *PublishNoteHandler {
	return &PublishNoteHandler{noteRepo: noteRepo}
}

var ProvidePublishNoteHandler = NewPublishNoteHandler

func (h *PublishNoteHandler) Handle(ctx context.Context, cmd *PublishNote) error {
	return nil
}
