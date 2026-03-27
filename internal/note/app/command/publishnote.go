package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
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

func (h *PublishNoteHandler) Handle(ctx context.Context, cmd *PublishNote) errs.Error {
	return nil
}
