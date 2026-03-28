package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type Folder struct {
	id              uuid.UUID
	name            string
	icon            *string
	workspaceID     uuid.UUID
	folderHierarchy FolderHierarchy
	trashed         *Trashed

	events []Event
}

func NewFolder(
	id uuid.UUID,
	name string,
	icon *string,
	workspaceID uuid.UUID,
	folderHierarchy FolderHierarchy,
) (*Folder, errs.Error) {
	if name == "" {
		return nil, errs.EmptyFolderName
	}
	folder := &Folder{
		id:              id,
		name:            name,
		icon:            icon,
		workspaceID:     workspaceID,
		folderHierarchy: folderHierarchy,
		trashed:         nil,

		events: []Event{},
	}
	folder.AddEvent(
		&FolderCreatedEvent{
			BaseEvent:   *NewBaseEvent(),
			AggregateID: folder.id,
			Name:        folder.name,
			Icon:        folder.icon,
		},
	)
	return folder, nil
}

func UnmarshalFolder(
	id uuid.UUID,
	name string,
	icon *string,
	workspaceID uuid.UUID,
	folderHierarchy FolderHierarchy,
	trashed *Trashed,
) *Folder {
	return &Folder{
		id:              id,
		name:            name,
		icon:            icon,
		workspaceID:     workspaceID,
		folderHierarchy: folderHierarchy,
		trashed:         trashed,

		events: []Event{},
	}
}

func (f *Folder) ID() uuid.UUID {
	return f.id
}

func (f *Folder) Name() string {
	return f.name
}

func (f *Folder) Rename(name string) {
	f.name = name
	f.AddEvent(&FolderUpdatedEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: f.id,
		Name:        f.name,
		Icon:        f.icon,
	})
}

func (f *Folder) Icon() *string {
	return f.icon
}

func (f *Folder) SetIcon(icon string) {
	f.icon = &icon
	f.AddEvent(&FolderUpdatedEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: f.id,
		Name:        f.name,
		Icon:        f.icon,
	})
}

func (f *Folder) WorkspaceID() uuid.UUID {
	return f.workspaceID
}

func (f *Folder) FolderHierarchy() FolderHierarchy {
	return f.folderHierarchy
}

func (f *Folder) ParentID() *uuid.UUID {
	return f.folderHierarchy.ParentID()
}

func (f *Folder) IsRoot() bool {
	return f.folderHierarchy.IsRoot()
}

func (f *Folder) MoveToFolder(folderID uuid.UUID) {
	hierarchy := NewFolderHierarchy(&folderID)
	f.folderHierarchy = *hierarchy
	f.AddEvent(
		&FolderMovedEvent{
			BaseEvent:   *NewBaseEvent(),
			AggregateID: f.id,
			ParentID:    folderID,
		},
	)
}

func (f *Folder) IsTrashed() bool {
	return f.trashed != nil
}

func (f *Folder) TrashedBy() *TrashedBy {
	if f.trashed == nil {
		return nil
	}
	return &f.trashed.by
}

func (f *Folder) TrashedByString() *string {
	if f.trashed == nil {
		return nil
	}
	return new(f.trashed.by.String())
}

func (f *Folder) TrashedAt() *time.Time {
	if f.trashed == nil {
		return nil
	}
	return &f.trashed.at
}

func (f *Folder) Trash(trashedBy TrashedBy) errs.Error {
	if f.trashed != nil {
		return errs.NewFolderAlreadyTrashed(f.id)
	}
	f.trashed = NewTrashed(trashedBy, time.Now())
	f.AddEvent(&FolderTrashedEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: f.id,
	})
	return nil
}

func (f *Folder) Restore() {
	f.trashed = nil
	f.AddEvent(&FolderRestoredEvent{
		BaseEvent:   *NewBaseEvent(),
		AggregateID: f.id,
	})
}

func (f *Folder) AddEvent(event Event) {
	f.events = append(f.events, event)
}

func (f *Folder) PopEvents() []Event {
	events := f.events
	f.events = []Event{}
	return events
}

type FolderHierarchy struct {
	parentID *uuid.UUID
}

func NewFolderHierarchy(parentID *uuid.UUID) *FolderHierarchy {
	return &FolderHierarchy{
		parentID: parentID,
	}
}

func (fh *FolderHierarchy) ParentID() *uuid.UUID {
	return fh.parentID
}

func (fh *FolderHierarchy) IsRoot() bool {
	return fh.parentID == nil
}
