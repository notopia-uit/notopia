package component

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

const DomainEventTopicPrefix = "events.internal.note."

func DomainEventToTopic(event domain.Event) (string, error) {
	switch event.(type) {
	case *domain.FolderCreatedEvent:
		return DomainEventTopicPrefix + "folder.created", nil
	case *domain.FolderDeletedEvent:
		return DomainEventTopicPrefix + "folder.deleted", nil
	case *domain.FolderRenamedEvent:
		return DomainEventTopicPrefix + "folder.renamed", nil
	case *domain.FolderIconChangedEvent:
		return DomainEventTopicPrefix + "folder.icon_changed", nil
	case *domain.FolderMovedEvent:
		return DomainEventTopicPrefix + "folder.moved", nil
	case *domain.FolderTrashedEvent:
		return DomainEventTopicPrefix + "folder.trashed", nil
	case *domain.FolderRestoredEvent:
		return DomainEventTopicPrefix + "folder.restored", nil
	case *domain.FolderPermanentlyDeletedEvent:
		return DomainEventTopicPrefix + "folder.permanently_deleted", nil
	case *domain.NoteCreatedEvent:
		return DomainEventTopicPrefix + "note.created", nil
	case *domain.NoteDeletedEvent:
		return DomainEventTopicPrefix + "note.deleted", nil
	case *domain.NoteRenamedEvent:
		return DomainEventTopicPrefix + "note.renamed", nil
	case *domain.NoteIconChangedEvent:
		return DomainEventTopicPrefix + "note.icon_changed", nil
	case *domain.NoteTagsChangedEvent:
		return DomainEventTopicPrefix + "note.tags_changed", nil
	case *domain.NoteSizeChangedEvent:
		return DomainEventTopicPrefix + "note.size_changed", nil
	case *domain.NoteMovedEvent:
		return DomainEventTopicPrefix + "note.moved", nil
	case *domain.NoteOutgoingLinksChangedEvent:
		return DomainEventTopicPrefix + "note.outgoing_links_changed", nil
	case *domain.NoteTrashedEvent:
		return DomainEventTopicPrefix + "note.trashed", nil
	case *domain.NoteRestoredEvent:
		return DomainEventTopicPrefix + "note.restored", nil
	case *domain.NotePermanentlyDeletedEvent:
		return DomainEventTopicPrefix + "note.permanently_deleted", nil
	case *domain.WorkspaceUpdatedEvent:
		return DomainEventTopicPrefix + "workspace.updated", nil
	case *domain.WorkspaceDeletedEvent:
		return DomainEventTopicPrefix + "workspace.deleted", nil
	default:
		return "", fmt.Errorf("unknown domain event type: %T", event)
	}
}

const IntegrationEventTopicPrefix = "events.integration.note."

func IntegrationEventToTopic(event app.IntegrationEvent) (string, error) {
	switch event.(type) {
	case app.IntegrationEventNoteCreated:
		return IntegrationEventTopicPrefix + "note.created", nil
	case app.IntegrationEventNoteDeleted:
		return IntegrationEventTopicPrefix + "note.deleted", nil
	case app.IntegrationEventNoteUpdated:
		return IntegrationEventTopicPrefix + "note.updated", nil
	}
	return "", fmt.Errorf("unknown integration event type: %T", event)
}
