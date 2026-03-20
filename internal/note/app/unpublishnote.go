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
	noterepo domain.NoteRepo
}

func NewUnpublishNoteHandler(noterepo domain.NoteRepo) *UnpublishNoteHandler {
	return &UnpublishNoteHandler{noterepo: noterepo}
}

var ProvideUnpublishNoteHandler = NewUnpublishNoteHandler

func (h *UnpublishNoteHandler) Handle(ctx context.Context, cmd *UnpublishNote) error {
	// TODO: domain.Note has no Unpublish() method. Add Unpublish() to domain.Note and a
	// published field, then call note.Unpublish() here before Save.
	_, err := h.noterepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return domain.NewErrNoteNotFound(cmd.ID, err)
	}
	return nil
}
