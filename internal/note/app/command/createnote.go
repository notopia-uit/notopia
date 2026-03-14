package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type CreateNote struct {
	ID       uuid.UUID
	Name     string
	Icon     *string
	Tags     []string
	FolderID uuid.UUID
}

type CreateNoteHandler struct {
	noterepo domain.NoteRepo
}

func NewCreateNoteHandler(noterepo domain.NoteRepo) *CreateNoteHandler {
	return &CreateNoteHandler{noterepo: noterepo}
}

func (h *CreateNoteHandler) Handle(ctx context.Context, cmd *CreateNote) error {
	note := domain.NewNote(cmd.ID, cmd.Name, cmd.Icon, cmd.Tags, cmd.FolderID)
	return h.noterepo.Save(ctx, note)
}
