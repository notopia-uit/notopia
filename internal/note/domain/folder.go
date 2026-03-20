package domain

import (
	"time"

	"github.com/google/uuid"
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
) (*Folder, error) {
	if name == "" {
		return nil, ErrEmptyFolderName
	}
	folder := &Folder{
		id:              id,
		name:            name,
		icon:            icon,
		workspaceID:     workspaceID,
		folderHierarchy: folderHierarchy,
	}
	folder.AddEvent(
		&FolderCreatedEvent{
			Id:   folder.id,
			Name: folder.name,
			Icon: folder.icon,
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
		ID:   f.id,
		Name: f.name,
		Icon: f.icon,
	})
}

func (f *Folder) Icon() *string {
	return f.icon
}

func (f *Folder) SetIcon(icon string) {
	f.icon = &icon
	f.AddEvent(&FolderUpdatedEvent{
		ID:   f.id,
		Name: f.name,
		Icon: f.icon,
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
			Id:       f.id,
			ParentId: folderID,
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

func (f *Folder) Trash(trashedBy TrashedBy) {
	f.trashed = NewTrashed(trashedBy, time.Now())
	f.AddEvent(&FolderTrashedEvent{
		Id: f.id,
	})
}

func (f *Folder) Restore() {
	f.trashed = nil
	f.AddEvent(&FolderRestoredEvent{
		Id: f.id,
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
