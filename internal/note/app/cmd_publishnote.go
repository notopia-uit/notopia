package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
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

type PublishNoteCmd commonhandler.Cmd[PublishNote]

var _ PublishNoteCmd = (*PublishNoteHandler)(nil)

func (h *PublishNoteHandler) Handle(ctx context.Context, cmd *PublishNote) error {
	return nil
}
