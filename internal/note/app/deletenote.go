package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type DeleteNote struct {
	ID uuid.UUID
}

type DeleteNoteHandler struct {
	noteRepo domain.NoteRepo
}

func NewDeleteNoteHandler(noteRepo domain.NoteRepo) *DeleteNoteHandler {
	return &DeleteNoteHandler{noteRepo: noteRepo}
}

var ProvideDeleteNoteHandler = NewDeleteNoteHandler

func (h *DeleteNoteHandler) Handle(ctx context.Context, cmd *DeleteNote) error {
	return h.noteRepo.PermanentlyDeleteByID(ctx, cmd.ID)
}
