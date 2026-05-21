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
	return nil
}
