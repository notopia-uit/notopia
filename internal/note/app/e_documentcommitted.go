package app

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type DocumentCommitted struct {
	Content         any
	ID              uuid.UUID
	OutgoingLinkIDs []uuid.UUID
	Tags            []string
	UserID          string
}

type DocumentCommittedHandler struct {
	noteRepo    domain.NoteRepo
	noteService *domain.UpdateNoteSizeService
}

func NewDocumentCommittedHandler(
	noteRepo domain.NoteRepo,
	noteService *domain.UpdateNoteSizeService,
) *DocumentCommittedHandler {
	return &DocumentCommittedHandler{
		noteRepo:    noteRepo,
		noteService: noteService,
	}
}

var ProvideDocumentCommittedHandler = NewDocumentCommittedHandler

func (h *DocumentCommittedHandler) Handle(ctx context.Context, event *DocumentCommitted) error {
	note, err := h.noteRepo.GetByID(ctx, event.ID, false)
	if err != nil {
		return err
	}
	err = h.noteService.Handle(note, event.Content, event.UserID)
	if err != nil {
		return err
	}
	note.SetTags(event.Tags, event.UserID)
	note.SetOutgoingLinks(event.OutgoingLinkIDs, event.UserID)
	slog.InfoContext(
		ctx,
		"Document committed event handled",
		slog.String("note_id", note.ID().String()),
		slog.Uint64("new_size", note.Size()),
	)
	return h.noteRepo.Save(ctx, note)
}
