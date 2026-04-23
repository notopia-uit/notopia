package app

import (
	"context"
	"log/slog"

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
	slog.DebugContext(ctx, "unpublishing note", slog.String("note_id", cmd.ID.String()))
	return nil
}
