package app

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

type WorkspaceEventPubSub interface {
	Publish(
		ctx context.Context,
		workspaceID uuid.UUID,
		userID string,
		events ...WorkspaceEvent,
	) errs.Error

	Subscribe(
		ctx context.Context,
		workspaceID uuid.UUID,
		userID string,
	) (<-chan WorkspaceEvent, errs.Error)

	Run(ctx context.Context) error

	Close() error

	Check(ctx context.Context) error
}

type WorkspaceEvent interface {
	IsWorkspaceEvent()
}

type WorkspaceEventWorkspaceItemsChanged struct {
	note.WorkspaceItemsUpdatedEvent
}

var _ WorkspaceEvent = (*WorkspaceEventWorkspaceItemsChanged)(nil)

func (e *WorkspaceEventWorkspaceItemsChanged) IsWorkspaceEvent() {}

type WorkspaceEventMembersUpdated struct {
	note.WorkspaceMembersUpdatedEvent
}

var _ WorkspaceEvent = (*WorkspaceEventMembersUpdated)(nil)

func (e *WorkspaceEventMembersUpdated) IsWorkspaceEvent() {}

type WorkspaceEventWorkspaceUpdated struct {
	note.WorkspaceUpdatedEvent
}

var _ WorkspaceEvent = (*WorkspaceEventWorkspaceUpdated)(nil)

func (e *WorkspaceEventWorkspaceUpdated) IsWorkspaceEvent() {}

type WorkspaceEventWorkspaceDeleted struct {
	note.WorkspaceDeletedEvent
}

var _ WorkspaceEvent = (*WorkspaceEventWorkspaceDeleted)(nil)

func (e *WorkspaceEventWorkspaceDeleted) IsWorkspaceEvent() {}

func FromDomainEventToWorkspaceEvent(event domain.Event) (WorkspaceEvent, bool) {
	switch e := event.(type) {
	case *domain.FolderCreatedEvent,
		*domain.FolderDeletedEvent,
		*domain.FolderUpdatedEvent,
		*domain.FolderMovedEvent,
		*domain.FolderTrashedEvent,
		*domain.FolderRestoredEvent,
		*domain.FolderPermanentlyDeletedEvent,
		*domain.NoteCreatedEvent,
		*domain.NoteDeletedEvent,
		*domain.NoteUpdatedEvent,
		*domain.NoteMovedEvent,
		*domain.NoteTrashedEvent,
		*domain.NoteRestoredEvent,
		*domain.NotePermanentlyDeletedEvent:
		return &WorkspaceEventWorkspaceItemsChanged{
			WorkspaceItemsUpdatedEvent: note.WorkspaceItemsUpdatedEvent{
				Type: note.WorkspaceItemsUpdatedEventTypeWorkspaceItemsUpdatedEvent,
				Data: note.WorkspaceItemsUpdatedEventData{
					WorkspaceId: (*note.PropertiesId)(new(e.GetAggregateID())),
				},
			},
		}, true
	case *domain.WorkspaceUpdatedEvent:
		return &WorkspaceEventWorkspaceUpdated{
			WorkspaceUpdatedEvent: note.WorkspaceUpdatedEvent{
				Type: note.WorkspaceUpdatedEventTypeWorkspaceUpdatedEvent,
				Data: note.Workspace{
					Id:   (*note.PropertiesId)(new(e.GetAggregateID())),
					Name: e.Name,
					Slug: e.Slug,
				},
			},
		}, true
	case *domain.WorkspaceDeletedEvent:
		return &WorkspaceEventWorkspaceDeleted{
			WorkspaceDeletedEvent: note.WorkspaceDeletedEvent{
				Type: note.WorkspaceDeletedEventTypeWorkspaceDeletedEvent,
				Data: note.WorkspaceDeletedEventData{
					Id: (*note.PropertiesId)(new(e.GetAggregateID())),
				},
			},
		}, true
	default:
		return nil, false
	}
}

var workspaceEventTypeRegistry = make(map[string]reflect.Type)

func init() {
	registerWorkspaceEventType(&WorkspaceEventWorkspaceItemsChanged{})
	registerWorkspaceEventType(&WorkspaceEventMembersUpdated{})
	registerWorkspaceEventType(&WorkspaceEventWorkspaceUpdated{})
	registerWorkspaceEventType(&WorkspaceEventWorkspaceDeleted{})
}

func registerWorkspaceEventType(event WorkspaceEvent) {
	eventType := reflect.TypeOf(event).Elem().Name()
	workspaceEventTypeRegistry[eventType] = reflect.TypeOf(event).Elem()
}

func NewEmptyWorkspaceEventFromType(t string) (WorkspaceEvent, bool) {
	if t, ok := workspaceEventTypeRegistry[t]; ok {
		return reflect.New(t).Interface().(WorkspaceEvent), true
	}
	return nil, false
}
