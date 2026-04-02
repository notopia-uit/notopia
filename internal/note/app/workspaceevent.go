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
	isWorkspaceEvent()
	GetID() uuid.UUID
	GetEvent() string
}

type workspaceEvent[E ~string] struct {
	Id    uuid.UUID `json:"id"`
	Event E         `json:"event"`
	Data  any       `json:"data"`
}

var _ WorkspaceEvent = (*workspaceEvent[string])(nil)

func (e workspaceEvent[E]) isWorkspaceEvent() {}
func (e workspaceEvent[E]) GetID() uuid.UUID  { return e.Id }
func (e workspaceEvent[E]) GetEvent() string  { return string(e.Event) }

type WorkspaceEventWorkspaceMembersUpdated struct {
	workspaceEvent[note.WorkspaceMembersUpdatedEventEvent]
}

type WorkspaceEventWorkspaceItemsChanged struct {
	workspaceEvent[note.WorkspaceItemsUpdatedEventEvent]
}

type WorkspaceEventMembersUpdated struct {
	workspaceEvent[note.WorkspaceMembersUpdatedEventEvent]
}

type WorkspaceEventWorkspaceUpdated struct {
	workspaceEvent[note.WorkspaceUpdatedEventEvent]
}

type WorkspaceEventWorkspaceDeleted struct {
	workspaceEvent[note.WorkspaceDeletedEventEvent]
}

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
			workspaceEvent: workspaceEvent[note.WorkspaceItemsUpdatedEventEvent]{
				Id:    e.GetID(),
				Event: note.WorkspaceItemsUpdatedEventEventWorkspaceItemsUpdatedEvent,
				Data: note.WorkspaceItemsUpdatedEventData{
					WorkspaceId: (*note.PropertiesId)(new(e.GetAggregateID())),
				},
			},
		}, true
	case *domain.WorkspaceUpdatedEvent:
		return &WorkspaceEventWorkspaceUpdated{
			workspaceEvent: workspaceEvent[note.WorkspaceUpdatedEventEvent]{
				Id:    e.GetID(),
				Event: note.WorkspaceUpdatedEventEventWorkspaceUpdatedEvent,
				Data: note.Workspace{
					Id:   (*note.PropertiesId)(new(e.GetAggregateID())),
					Name: e.Name,
					Slug: e.Slug,
				},
			},
		}, true
	case *domain.WorkspaceDeletedEvent:
		return &WorkspaceEventWorkspaceDeleted{
			workspaceEvent: workspaceEvent[note.WorkspaceDeletedEventEvent]{
				Id:    e.GetID(),
				Event: note.WorkspaceDeletedEventEventWorkspaceDeletedEvent,
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
	registerWorkspaceEventType(
		//exhaustruct:ignore
		&WorkspaceEventWorkspaceMembersUpdated{},
	)
	registerWorkspaceEventType(
		//exhaustruct:ignore
		&WorkspaceEventWorkspaceItemsChanged{},
	)
	registerWorkspaceEventType(
		//exhaustruct:ignore
		&WorkspaceEventMembersUpdated{},
	)
	registerWorkspaceEventType(
		//exhaustruct:ignore
		&WorkspaceEventWorkspaceUpdated{},
	)
	registerWorkspaceEventType(
		//exhaustruct:ignore
		&WorkspaceEventWorkspaceDeleted{},
	)
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
