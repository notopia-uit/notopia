package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishNote struct {
	ID uuid.UUID
}

type PublishNoteHandler struct {
	noterepo domain.NoteRepo
}

func NewPublishNoteHandler(noterepo domain.NoteRepo) *PublishNoteHandler {
	return &PublishNoteHandler{noterepo: noterepo}
}

func (h *PublishNoteHandler) Handle(ctx context.Context, cmd *PublishNote) error {
	// TODO: domain.Note has no Publish() method. Add Publish() to domain.Note and a
	// published field, then call note.Publish() here before Save.
	_, err := h.noterepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return domain.NewErrNoteNotFound(cmd.ID, err)
	}
	return nil
}
