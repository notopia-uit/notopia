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
	ID   uuid.UUID
	Name string
	Icon string
}

func (e IntegrationEventNoteCreated) isIntegrationEvent() {}

type IntegrationEventNoteDeleted struct {
	ID uuid.UUID
}

func (e IntegrationEventNoteDeleted) isIntegrationEvent() {}

type IntegrationEventNoteUpdated struct {
	ID            uuid.UUID
	Name          string
	Icon          string
	Tags          []string
	Size          uint64
	FolderID      uuid.UUID
	OutgoingLinks uuid.UUIDs
	UpdatedAt     time.Time
}

func (e IntegrationEventNoteUpdated) isIntegrationEvent() {}

type UserWorkspaceRoleUpdated struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Role        WorkspaceRole // unspecified will be role lost
}
