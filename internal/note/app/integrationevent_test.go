package app

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Interface implementation tests
// ─────────────────────────────────────────────

func TestIntegrationEventNoteCreated_ImplementsInterface(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	event := IntegrationEventNoteCreated{
		ID:   id,
		Name: "Test Note",
		Icon: "📝",
	}

	// Verify it satisfies the IntegrationEvent interface at compile time
	var _ IntegrationEvent = event

	// Verify the marker method does not panic
	assert.NotPanics(t, func() { event.isIntegrationEvent() })
}

func TestIntegrationEventNoteDeleted_ImplementsInterface(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	event := IntegrationEventNoteDeleted{
		ID: id,
	}

	var _ IntegrationEvent = event

	assert.NotPanics(t, func() { event.isIntegrationEvent() })
}

func TestIntegrationEventNoteUpdated_ImplementsInterface(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	folderID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	linkID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	now := time.Now()

	event := IntegrationEventNoteUpdated{
		ID:            id,
		Name:          "Updated Note",
		Icon:          "✏️",
		Tags:          []string{"tag1", "tag2"},
		Size:          1024,
		FolderID:      folderID,
		OutgoingLinks: uuid.UUIDs{linkID},
		UpdatedAt:     now,
	}

	var _ IntegrationEvent = event

	assert.NotPanics(t, func() { event.isIntegrationEvent() })
}

// ─────────────────────────────────────────────
// IntegrationEventNoteCreated field tests
// ─────────────────────────────────────────────

func TestIntegrationEventNoteCreated_Fields(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	event := IntegrationEventNoteCreated{
		ID:   id,
		Name: "My Note",
		Icon: "🗒",
	}

	assert.Equal(t, id, event.ID)
	assert.Equal(t, "My Note", event.Name)
	assert.Equal(t, "🗒", event.Icon)
}

func TestIntegrationEventNoteCreated_ZeroValues(t *testing.T) {
	t.Parallel()

	event := IntegrationEventNoteCreated{}

	assert.Equal(t, uuid.Nil, event.ID)
	assert.Equal(t, "", event.Name)
	assert.Equal(t, "", event.Icon)
}

// ─────────────────────────────────────────────
// IntegrationEventNoteDeleted field tests
// ─────────────────────────────────────────────

func TestIntegrationEventNoteDeleted_Fields(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	event := IntegrationEventNoteDeleted{ID: id}

	assert.Equal(t, id, event.ID)
}

// ─────────────────────────────────────────────
// IntegrationEventNoteUpdated field tests
// ─────────────────────────────────────────────

func TestIntegrationEventNoteUpdated_Fields(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	folderID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	link1 := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	updatedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	event := IntegrationEventNoteUpdated{
		ID:            id,
		Name:          "Note Updated",
		Icon:          "📔",
		Tags:          []string{"a", "b", "c"},
		Size:          512,
		FolderID:      folderID,
		OutgoingLinks: uuid.UUIDs{link1},
		UpdatedAt:     updatedAt,
	}

	require.Equal(t, id, event.ID)
	assert.Equal(t, "Note Updated", event.Name)
	assert.Equal(t, "📔", event.Icon)
	assert.Equal(t, []string{"a", "b", "c"}, event.Tags)
	assert.Equal(t, uint64(512), event.Size)
	assert.Equal(t, folderID, event.FolderID)
	assert.Equal(t, uuid.UUIDs{link1}, event.OutgoingLinks)
	assert.Equal(t, updatedAt, event.UpdatedAt)
}

// ─────────────────────────────────────────────
// IntegrationPublisher interface usage
// ─────────────────────────────────────────────

func TestIntegrationPublisher_AcceptsMultipleEventTypes(t *testing.T) {
	t.Parallel()

	publisher := &mockIntegrationPublisher{}

	events := []IntegrationEvent{
		IntegrationEventNoteCreated{ID: uuid.New(), Name: "Created", Icon: ""},
		IntegrationEventNoteDeleted{ID: uuid.New()},
		IntegrationEventNoteUpdated{ID: uuid.New(), Name: "Updated"},
	}

	ctx := t.Context()
	err := publisher.Publish(ctx, events...)

	require.NoError(t, err)
	assert.Len(t, publisher.publishedEvents, 3)
}