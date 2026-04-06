package app

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/share"
)

type IntegrationPublisher interface {
	Publish(ctx context.Context, event ...IntegrationEvent) error
}

type IntegrationEventType string

func (t IntegrationEventType) String() string {
	return string(t)
}

var (
	IntegrationEventTypeNoteCreated IntegrationEventType = "NoteCreated"
	IntegrationEventTypeNoteDeleted IntegrationEventType = "NoteDeleted"
	IntegrationEventTypeNoteUpdated IntegrationEventType = "NoteUpdated"
)

type IntegrationEvent interface {
	isIntegrationEvent()
	Type() IntegrationEventType
}
type IntegrationEventNoteCreated share.NoteCreatedEvent

var _ IntegrationEvent = (*IntegrationEventNoteCreated)(nil)

func (e IntegrationEventNoteCreated) isIntegrationEvent() {}

func (e IntegrationEventNoteCreated) Type() IntegrationEventType {
	return IntegrationEventTypeNoteCreated
}

type IntegrationEventNoteDeleted share.NoteDeletedEvent

var _ IntegrationEvent = (*IntegrationEventNoteDeleted)(nil)

func (e IntegrationEventNoteDeleted) isIntegrationEvent() {}

func (e IntegrationEventNoteDeleted) Type() IntegrationEventType {
	return IntegrationEventTypeNoteDeleted
}

type IntegrationEventNoteUpdated share.NoteUpdatedEvent

var _ IntegrationEvent = (*IntegrationEventNoteUpdated)(nil)

func (e IntegrationEventNoteUpdated) isIntegrationEvent() {}

func (e IntegrationEventNoteUpdated) Type() IntegrationEventType {
	return IntegrationEventTypeNoteUpdated
}
