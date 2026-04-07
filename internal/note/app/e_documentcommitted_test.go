package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentCommittedHandler(t *testing.T) {
	t.Parallel()

	noteRepo := &mockNoteRepo{}
	noteService := domain.NewNoteService()

	handler := NewDocumentCommittedHandler(noteRepo, noteService)

	require.NotNil(t, handler)
}

func TestDocumentCommittedHandler_Handle_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noteID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	folderID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	note := domain.NewNote(noteID, "Test Note", "", folderID)

	noteRepo := &mockNoteRepo{note: note}
	noteService := domain.NewNoteService()
	handler := NewDocumentCommittedHandler(noteRepo, noteService)

	outgoingLinkID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	event := &DocumentCommitted{
		ID:              noteID,
		Content:         map[string]any{"text": "hello world"},
		Tags:            []string{"work", "meeting"},
		OutgoingLinkIDs: []uuid.UUID{outgoingLinkID},
		UserID:          "user-001",
	}

	err := handler.Handle(ctx, event)

	require.NoError(t, err)
	assert.Equal(t, []string{"work", "meeting"}, note.Tags())
	assert.Equal(t, uuid.UUIDs{outgoingLinkID}, note.OutgoingLinks())
}

func TestDocumentCommittedHandler_Handle_NoteRepoGetByIDError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteRepo := &mockNoteRepo{getByIDErr: errSentinel}
	noteService := domain.NewNoteService()
	handler := NewDocumentCommittedHandler(noteRepo, noteService)

	event := &DocumentCommitted{
		ID:     uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
		UserID: "user-002",
	}

	err := handler.Handle(ctx, event)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
}

func TestDocumentCommittedHandler_Handle_UpdateSizeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noteID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	folderID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	note := domain.NewNote(noteID, "Test Note", "", folderID)

	noteRepo := &mockNoteRepo{note: note}
	noteService := domain.NewNoteService()
	handler := NewDocumentCommittedHandler(noteRepo, noteService)

	// A channel is not JSON-serializable, which should cause UpdateNoteSizeBasedOnContent to fail
	event := &DocumentCommitted{
		ID:      noteID,
		Content: make(chan int),
		UserID:  "user-003",
	}

	err := handler.Handle(ctx, event)

	require.Error(t, err, "should fail to marshal channel content")
}

func TestDocumentCommittedHandler_Handle_SaveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noteID := uuid.MustParse("11111111-aaaa-aaaa-aaaa-111111111111")
	folderID := uuid.MustParse("22222222-bbbb-bbbb-bbbb-222222222222")
	note := domain.NewNote(noteID, "Save Error Note", "", folderID)

	noteRepo := &mockNoteRepo{note: note, saveErr: errSentinel}
	noteService := domain.NewNoteService()
	handler := NewDocumentCommittedHandler(noteRepo, noteService)

	event := &DocumentCommitted{
		ID:      noteID,
		Content: "plain string content",
		Tags:    []string{"tag1"},
		UserID:  "user-004",
	}

	err := handler.Handle(ctx, event)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
}

func TestDocumentCommittedHandler_Handle_EmptyTagsAndLinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noteID := uuid.MustParse("33333333-cccc-cccc-cccc-333333333333")
	folderID := uuid.MustParse("44444444-dddd-dddd-dddd-444444444444")
	note := domain.NewNote(noteID, "Empty Test Note", "🗒", folderID)

	noteRepo := &mockNoteRepo{note: note}
	noteService := domain.NewNoteService()
	handler := NewDocumentCommittedHandler(noteRepo, noteService)

	event := &DocumentCommitted{
		ID:              noteID,
		Content:         nil,
		Tags:            []string{},
		OutgoingLinkIDs: []uuid.UUID{},
		UserID:          "user-005",
	}

	err := handler.Handle(ctx, event)

	require.NoError(t, err)
	assert.Empty(t, note.Tags())
	assert.Empty(t, note.OutgoingLinks())
}