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
	noterepo domain.NoteRepo
}

func NewDeleteNoteHandler(noterepo domain.NoteRepo) *DeleteNoteHandler {
	return &DeleteNoteHandler{noterepo: noterepo}
}

var ProvideDeleteNoteHandler = NewDeleteNoteHandler

func (h *DeleteNoteHandler) Handle(ctx context.Context, cmd *DeleteNote) error {
	return h.noterepo.PermanentlyDeleteByID(ctx, cmd.ID)
}
