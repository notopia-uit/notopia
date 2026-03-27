package event

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

type DocumentCommitted share.DocumentCommittedEvent

type DocumentCommittedHandler struct {
	noteRepo    domain.NoteRepo
	noteService *domain.NoteService
}

func NewDocumentCommittedHandler(
	noteRepo domain.NoteRepo,
	noteService *domain.NoteService,
) *DocumentCommittedHandler {
	return &DocumentCommittedHandler{
		noteRepo:    noteRepo,
		noteService: noteService,
	}
}

var ProvideDocumentCommittedHandler = NewDocumentCommittedHandler

func (h *DocumentCommittedHandler) Handle(ctx context.Context, event *DocumentCommitted) error {
	note, err := h.noteRepo.GetByID(ctx, event.Id, false)
	if err != nil {
		return err
	}
	err = h.noteService.UpdateNoteSizeBasedOnContent(note, event.Content)
	if err != nil {
		return err
	}
	note.SetTags(event.Tags)
	note.SetOutgoingLinks(event.OutgoingLinkIds)
	slog.InfoContext(
		ctx,
		"Document committed event handled",
		slog.String("note_id", note.ID().String()),
		slog.Uint64("new_size", note.Size()),
	)
	return h.noteRepo.Save(ctx, note)
}
