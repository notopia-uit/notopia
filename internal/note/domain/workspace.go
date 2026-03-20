package domain

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	id           uuid.UUID
	name         string
	slug         string
	rootFolderID uuid.UUID
	deletedAt    *time.Time

	event []Event
}

func NewWorkspace(
	id uuid.UUID,
	name string,
	slug string,
	rootFolderID uuid.UUID,
) *Workspace {
	return &Workspace{
		id:           id,
		name:         name,
		slug:         slug,
		rootFolderID: rootFolderID,

		event: []Event{},
	}
}

func (w *Workspace) ID() uuid.UUID {
	return w.id
}

func (w *Workspace) Name() string {
	return w.name
}

func (w *Workspace) Rename(name string) {
	w.name = name
	w.AddEvent(&WorkspaceUpdatedEvent{
		Id:   w.id,
		Name: w.name,
		Slug: w.slug,
	})
}

func (w *Workspace) Slug() string {
	return w.slug
}

func (w *Workspace) RootFolderID() uuid.UUID {
	return w.rootFolderID
}

func (w *Workspace) DeletedAt() *time.Time {
	return w.deletedAt
}

func (w *Workspace) Delete() {
	w.deletedAt = new(time.Now())
	w.AddEvent(&WorkspaceDeletedEvent{
		Id: w.id,
	})
}

func (w *Workspace) AddEvent(event Event) {
	w.event = append(w.event, event)
}

func (w *Workspace) PopEvents() []Event {
	events := w.event
	w.event = []Event{}
	return events
}
