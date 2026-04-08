package component

import "github.com/notopia-uit/notopia/internal/note/domain"

const DomainEventTopicPrefix = "events.internal.note."

func DomainEventToTopic(event domain.Event) (string, bool) {
	eType := domain.GetEventType(event)
	switch eType {
	case domain.EventTypeFolderCreated:
		return DomainEventTopicPrefix + "folder.created", true
	case domain.EventTypeFolderDeleted:
		return DomainEventTopicPrefix + "folder.deleted", true
	case domain.EventTypeFolderUpdated:
		return DomainEventTopicPrefix + "folder.updated", true
	case domain.EventTypeFolderMoved:
		return DomainEventTopicPrefix + "folder.moved", true
	case domain.EventTypeFolderTrashed:
		return DomainEventTopicPrefix + "folder.trashed", true
	case domain.EventTypeFolderRestored:
		return DomainEventTopicPrefix + "folder.restored", true
	case domain.EventTypeFolderPermanentlyDeleted:
		return DomainEventTopicPrefix + "folder.permanently_deleted", true
	case domain.EventTypeNoteCreated:
		return DomainEventTopicPrefix + "note.created", true
	case domain.EventTypeNoteDeleted:
		return DomainEventTopicPrefix + "note.deleted", true
	case domain.EventTypeNoteUpdated:
		return DomainEventTopicPrefix + "note.updated", true
	case domain.EventTypeNoteMoved:
		return DomainEventTopicPrefix + "note.moved", true
	case domain.EventTypeNoteTrashed:
		return DomainEventTopicPrefix + "note.trashed", true
	case domain.EventTypeNoteRestored:
		return DomainEventTopicPrefix + "note.restored", true
	case domain.EventTypeNotePermanentlyDeleted:
		return DomainEventTopicPrefix + "note.permanently_deleted", true
	case domain.EventTypeWorkspaceUpdated:
		return DomainEventTopicPrefix + "workspace.updated", true
	case domain.EventTypeWorkspaceDeleted:
		return DomainEventTopicPrefix + "workspace.deleted", true
	case domain.EventTypeUnspecified:
		return "", false
	default:
		return "", false
	}
}
