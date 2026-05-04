package app

import (
	"context"
	"log/slog"

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
	slog.DebugContext(
		ctx, "publishing note",
		slog.String("note_id", cmd.ID.String()),
		slog.String("user_id", cmd.UserID),
	)
	return nil
}
