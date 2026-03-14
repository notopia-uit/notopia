package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type RenameNote struct {
	ID   uuid.UUID
	Name string
}

type RenameNoteHandler struct {
	noterepo domain.NoteRepo
}

func NewRenameNoteHandler(noterepo domain.NoteRepo) *RenameNoteHandler {
	return &RenameNoteHandler{noterepo: noterepo}
}

func (h *RenameNoteHandler) Handle(ctx context.Context, cmd *RenameNote) error {
	note, err := h.noterepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return domain.NewErrNoteNotFound(cmd.ID, err)
	}
	note.Rename(cmd.Name)
	return h.noterepo.Save(ctx, note)
}
