package app

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotifyWorkspaceItemsUpdatedHandler(t *testing.T) {
	t.Parallel()

	publisher := &mockWorkspaceEventPublisher{}
	handler := NewNotifyWorkspaceItemsUpdatedHandler(publisher)

	require.NotNil(t, handler)
	assert.Equal(t, 1*time.Second, handler.debounceDuration)
}

func TestNotifyWorkspaceItemsUpdatedType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, NotifyWorkspaceItemsUpdatedType(0), NotifyWorkspaceItemsUpdatedTypeFolder)
	assert.Equal(t, NotifyWorkspaceItemsUpdatedType(1), NotifyWorkspaceItemsUpdatedTypeNote)
}

func TestNotifyWorkspaceItemsUpdatedHandler_Handle_PublishesAfterDebounce(t *testing.T) {
	t.Parallel()

	publisher := &mockWorkspaceEventPublisher{}
	handler := NewNotifyWorkspaceItemsUpdatedHandler(publisher)
	// Use a much shorter debounce for test speed
	handler.debounceDuration = 20 * time.Millisecond

	workspaceID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	itemID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	params := &NotifyWorkspaceItemsUpdated{
		UserID:          "user-notify",
		WorkspaceID:     workspaceID,
		WorkspaceItemID: itemID,
		Type:            NotifyWorkspaceItemsUpdatedTypeNote,
	}

	handler.Handle(params)

	// Wait for the debounce to fire (20ms) plus a buffer
	time.Sleep(60 * time.Millisecond)

	assert.Equal(t, 1, publisher.callCount, "publisher should have been called once after debounce")
	require.Len(t, publisher.publishedIDs, 1)
	assert.Equal(t, workspaceID, publisher.publishedIDs[0])
}

func TestNotifyWorkspaceItemsUpdatedHandler_Handle_DebouncesMultipleCalls(t *testing.T) {
	t.Parallel()

	publisher := &mockWorkspaceEventPublisher{}
	handler := NewNotifyWorkspaceItemsUpdatedHandler(publisher)
	handler.debounceDuration = 50 * time.Millisecond

	workspaceID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	params := &NotifyWorkspaceItemsUpdated{
		UserID:      "user-debounce",
		WorkspaceID: workspaceID,
		Type:        NotifyWorkspaceItemsUpdatedTypeNote,
	}

	// Fire multiple calls rapidly — they should be debounced to a single publish
	for i := 0; i < 5; i++ {
		handler.Handle(params)
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for the debounce to fire
	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, 1, publisher.callCount, "rapid calls should be debounced into one publish")
}

func TestNotifyWorkspaceItemsUpdatedHandler_Handle_DifferentWorkspacesDebouncedSeparately(t *testing.T) {
	t.Parallel()

	publisher := &mockWorkspaceEventPublisher{}
	handler := NewNotifyWorkspaceItemsUpdatedHandler(publisher)
	handler.debounceDuration = 20 * time.Millisecond

	workspaceA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-000000000001")
	workspaceB := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-000000000002")

	handler.Handle(&NotifyWorkspaceItemsUpdated{
		UserID:      "user-a",
		WorkspaceID: workspaceA,
		Type:        NotifyWorkspaceItemsUpdatedTypeFolder,
	})
	handler.Handle(&NotifyWorkspaceItemsUpdated{
		UserID:      "user-b",
		WorkspaceID: workspaceB,
		Type:        NotifyWorkspaceItemsUpdatedTypeNote,
	})

	// Wait for both debounces to fire
	time.Sleep(80 * time.Millisecond)

	assert.Equal(t, 2, publisher.callCount, "each workspace should produce its own publish")
}

func TestNotifyWorkspaceItemsUpdatedHandler_Handle_PublisherError_LogsAndContinues(t *testing.T) {
	t.Parallel()

	publisher := &mockWorkspaceEventPublisher{err: errSentinel}
	handler := NewNotifyWorkspaceItemsUpdatedHandler(publisher)
	handler.debounceDuration = 20 * time.Millisecond

	workspaceID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	params := &NotifyWorkspaceItemsUpdated{
		UserID:      "user-err",
		WorkspaceID: workspaceID,
		Type:        NotifyWorkspaceItemsUpdatedTypeNote,
	}

	// Handle should not panic even when publisher returns error
	assert.NotPanics(t, func() {
		handler.Handle(params)
		time.Sleep(60 * time.Millisecond)
	})

	// Publisher was called but returned an error; debouncer entry is cleaned up
	assert.Equal(t, 1, publisher.callCount, "publisher should be called even if it errors")
}

func TestNotifyWorkspaceItemsUpdatedHandler_Handle_DebounceEntryCleanedUpAfterFire(t *testing.T) {
	t.Parallel()

	publisher := &mockWorkspaceEventPublisher{}
	handler := NewNotifyWorkspaceItemsUpdatedHandler(publisher)
	handler.debounceDuration = 20 * time.Millisecond

	workspaceID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	handler.Handle(&NotifyWorkspaceItemsUpdated{
		UserID:      "user-cleanup",
		WorkspaceID: workspaceID,
		Type:        NotifyWorkspaceItemsUpdatedTypeNote,
	})

	time.Sleep(60 * time.Millisecond)

	// After first fire the debouncer should have been deleted
	_, loaded := handler.debouncers.Load(workspaceID)
	assert.False(t, loaded, "debouncer entry should be removed after firing")
}