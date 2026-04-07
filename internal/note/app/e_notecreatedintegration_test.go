package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoteCreatedDomainToIntegrationEventHandler(t *testing.T) {
	t.Parallel()

	publisher := &mockIntegrationPublisher{}
	noteRepo := &mockNoteRepo{}

	handler := NewNoteCreatedDomainToIntegrationEventHandler(publisher, noteRepo)

	require.NotNil(t, handler)
}

func TestNoteCreatedDomainToIntegrationEventHandler_Handle_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publisher := &mockIntegrationPublisher{}
	noteRepo := &mockNoteRepo{}
	handler := NewNoteCreatedDomainToIntegrationEventHandler(publisher, noteRepo)

	noteID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := "user-123"

	event := &domain.NoteCreatedEvent{
		BaseEvent: domain.NewBaseEvent(noteID, userID),
		Name:      "My Test Note",
		Icon:      "📝",
	}

	err := handler.Handle(ctx, event)

	require.NoError(t, err)
	require.Len(t, publisher.publishedEvents, 1)

	published, ok := publisher.publishedEvents[0].(IntegrationEventNoteCreated)
	require.True(t, ok, "expected IntegrationEventNoteCreated")
	assert.Equal(t, noteID, published.ID)
	assert.Equal(t, "My Test Note", published.Name)
	assert.Equal(t, "📝", published.Icon)
}

func TestNoteCreatedDomainToIntegrationEventHandler_Handle_PublisherError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publisher := &mockIntegrationPublisher{err: errSentinel}
	noteRepo := &mockNoteRepo{}
	handler := NewNoteCreatedDomainToIntegrationEventHandler(publisher, noteRepo)

	noteID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := "user-456"

	event := &domain.NoteCreatedEvent{
		BaseEvent: domain.NewBaseEvent(noteID, userID),
		Name:      "Another Note",
		Icon:      "",
	}

	err := handler.Handle(ctx, event)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel, "error should wrap the publisher error")
	assert.Len(t, publisher.publishedEvents, 0, "no events should be recorded on failure")
}

func TestNoteCreatedDomainToIntegrationEventHandler_Handle_EmptyIcon(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publisher := &mockIntegrationPublisher{}
	noteRepo := &mockNoteRepo{}
	handler := NewNoteCreatedDomainToIntegrationEventHandler(publisher, noteRepo)

	noteID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	userID := "user-789"

	event := &domain.NoteCreatedEvent{
		BaseEvent: domain.NewBaseEvent(noteID, userID),
		Name:      "Untitled Note",
		Icon:      "",
	}

	err := handler.Handle(ctx, event)

	require.NoError(t, err)
	require.Len(t, publisher.publishedEvents, 1)

	published, ok := publisher.publishedEvents[0].(IntegrationEventNoteCreated)
	require.True(t, ok)
	assert.Equal(t, "", published.Icon)
}