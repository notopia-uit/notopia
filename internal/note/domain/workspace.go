package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type Workspace struct {
	id        uuid.UUID
	name      string
	slug      string
	deletedAt *time.Time

	event []Event
}

func NewWorkspace(
	id uuid.UUID,
	name string,
	slug string,
) (*Workspace, error) {
	if name == "" {
		return nil, errs.EmptyFolderName
	}
	if slug == "" {
		return nil, errs.InvalidWorkspaceSlug
	}
	return &Workspace{
		id:        id,
		name:      name,
		slug:      slug,
		deletedAt: nil,

		event: []Event{},
	}, nil
}

func UnmarshalWorkspace(
	id uuid.UUID,
	name string,
	slug string,
	deletedAt *time.Time,
) *Workspace {
	return &Workspace{
		id:        id,
		name:      name,
		slug:      slug,
		deletedAt: deletedAt,

		event: []Event{},
	}
}

func (w *Workspace) ID() uuid.UUID {
	return w.id
}

func (w *Workspace) Name() string {
	return w.name
}

func (w *Workspace) Rename(name string, userID string) {
	w.name = name
	w.addEvent(&WorkspaceUpdatedEvent{
		BaseEvent: NewBaseEvent(w.id, userID),
		Name:      w.name,
		Slug:      w.slug,
	})
}

func (w *Workspace) Slug() string {
	return w.slug
}

func (w *Workspace) DeletedAt() *time.Time {
	return w.deletedAt
}

func (w *Workspace) Delete(userID string) {
	w.deletedAt = new(time.Now())
	w.addEvent(&WorkspaceDeletedEvent{
		BaseEvent: NewBaseEvent(w.id, userID),
	})
}

func (w *Workspace) addEvent(event Event) {
	w.event = append(w.event, event)
}

func (w *Workspace) PopEvents() []Event {
	events := w.event
	w.event = []Event{}
	return events
}
