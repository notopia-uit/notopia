package app

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

type DocumentCommitted share.DocumentCommittedEvent

type DocumentCommittedHandler struct {
	noterepo    domain.NoteRepo
	noteService *domain.NoteService
}

func NewDocumentCommittedHandler(
	noterepo domain.NoteRepo,
	noteService *domain.NoteService,
) *DocumentCommittedHandler {
	return &DocumentCommittedHandler{
		noterepo:    noterepo,
		noteService: noteService,
	}
}

var ProvideDocumentCommittedHandler = NewDocumentCommittedHandler

func (h *DocumentCommittedHandler) Handle(ctx context.Context, event *DocumentCommitted) error {
	note, err := h.noterepo.GetByID(ctx, event.Id, false)
	if err != nil {
		return err
	}
	err = h.noteService.UpdateNoteSizeBasedOnContent(note, event.Content)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Document committed event handled", "note_id", note.ID, "new_size", note.Size)
	return h.noterepo.Save(ctx, note)
}
