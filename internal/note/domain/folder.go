package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type Folder struct {
	id              uuid.UUID
	name            string
	icon            string
	workspaceID     uuid.UUID
	folderHierarchy FolderHierarchy
	trashed         *Trashed
	deleted         bool

	events []Event
}

func NewFolder(
	id uuid.UUID,
	name string,
	icon string,
	workspaceID uuid.UUID,
	folderHierarchy FolderHierarchy,
	userID string,
) (*Folder, error) {
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
		deleted:         false,

		events: []Event{},
	}
	folder.addEvent(
		&FolderCreatedEvent{
			BaseEvent: NewBaseEvent(folder.id, userID),
			Name:      folder.name,
			Icon:      folder.icon,
		},
	)
	return folder, nil
}

func UnmarshalFolder(
	id uuid.UUID,
	name string,
	icon string,
	workspaceID uuid.UUID,
	folderHierarchy FolderHierarchy,
	trashed *Trashed,
	deleted bool,
) *Folder {
	return &Folder{
		id:              id,
		name:            name,
		icon:            icon,
		workspaceID:     workspaceID,
		folderHierarchy: folderHierarchy,
		trashed:         trashed,
		deleted:         deleted,

		events: []Event{},
	}
}

func (f *Folder) ID() uuid.UUID {
	return f.id
}

func (f *Folder) Name() string {
	return f.name
}

func (f *Folder) Rename(name string, userID string) {
	f.name = name
	f.addEvent(&FolderUpdatedEvent{
		BaseEvent: NewBaseEvent(f.id, userID),
		Name:      f.name,
		Icon:      f.icon,
	})
}

func (f *Folder) Icon() string {
	return f.icon
}

func (f *Folder) SetIcon(icon string, userID string) {
	f.icon = icon
	f.addEvent(&FolderUpdatedEvent{
		BaseEvent: NewBaseEvent(f.id, userID),
		Name:      f.name,
		Icon:      f.icon,
	})
}

func (f *Folder) WorkspaceID() uuid.UUID {
	return f.workspaceID
}

func (f *Folder) FolderHierarchy() FolderHierarchy {
	return f.folderHierarchy
}

func (f *Folder) ParentID() uuid.UUID {
	return f.folderHierarchy.ParentID()
}

func (f *Folder) IsRoot() bool {
	return f.folderHierarchy.IsRoot()
}

func (f *Folder) MoveToFolder(folderID uuid.UUID, userID string) {
	f.folderHierarchy = NewFolderHierarchy(folderID)
	f.addEvent(
		&FolderMovedEvent{
			BaseEvent: NewBaseEvent(f.id, userID),
			ParentID:  folderID,
		},
	)
}

func (f *Folder) IsTrashed() bool {
	return f.trashed != nil
}

func (f *Folder) TrashedBy() TrashedBy {
	return f.trashed.By()
}

func (f *Folder) TrashedAt() time.Time {
	return f.trashed.At()
}

func (f *Folder) Trash(trashedBy TrashedBy, userID string) error {
	if f.trashed != nil {
		return errs.NewFolderAlreadyTrashed(f.id)
	}
	f.trashed = NewTrashed(trashedBy, time.Now())
	f.addEvent(&FolderTrashedEvent{
		BaseEvent: NewBaseEvent(f.id, userID),
	})
	return nil
}

func (f *Folder) Restore(userID string) {
	f.trashed = nil
	f.addEvent(&FolderRestoredEvent{
		BaseEvent: NewBaseEvent(f.id, userID),
	})
}

func (f *Folder) Deleted() bool {
	return f.deleted
}

func (f *Folder) PermanentlyDelete(userID string) {
	f.deleted = true
	f.addEvent(&FolderPermanentlyDeletedEvent{
		BaseEvent: NewBaseEvent(f.id, userID),
	})
}

func (f *Folder) addEvent(event Event) {
	f.events = append(f.events, event)
}

func (f *Folder) PopEvents() []Event {
	events := f.events
	f.events = []Event{}
	return events
}

type FolderHierarchy struct {
	parentID uuid.UUID
}

func NewFolderHierarchy(parentID uuid.UUID) FolderHierarchy {
	return FolderHierarchy{
		parentID: parentID,
	}
}

func (fh *FolderHierarchy) ParentID() uuid.UUID {
	return fh.parentID
}

func (fh *FolderHierarchy) IsRoot() bool {
	return fh.parentID == uuid.Nil
}
