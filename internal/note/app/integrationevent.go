package app

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TODO: Hey quite miss the user ID right?, but it should need when other service need it ye...

type IntegrationPublisher interface {
	Publish(ctx context.Context, event ...IntegrationEvent) error
}

type IntegrationEvent interface {
	isIntegrationEvent()
}

type IntegrationEventNoteCreated struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Icon        string
}

type IntegrationEventNoteDeleted struct {
	ID uuid.UUID
}

type IntegrationEventNoteUpdated struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Icon        string
	Size        uint64
	FolderID    uuid.UUID
	FolderName  string
	Trashed     Trashed
	UpdatedAt   time.Time
}

var (
	_ IntegrationEvent = (*IntegrationEventNoteCreated)(nil)
	_ IntegrationEvent = (*IntegrationEventNoteDeleted)(nil)
	_ IntegrationEvent = (*IntegrationEventNoteUpdated)(nil)
)

func (e IntegrationEventNoteCreated) isIntegrationEvent() {}
func (e IntegrationEventNoteDeleted) isIntegrationEvent() {}
func (e IntegrationEventNoteUpdated) isIntegrationEvent() {}
