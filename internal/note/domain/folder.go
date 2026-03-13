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
	trashedBy       *TrashedBy
	trashedAt       *time.Time
}

func NewFolder(
	id uuid.UUID,
	name string,
	icon *string,
	workspaceID uuid.UUID,
	folderHierarchy FolderHierarchy,
) *Folder {
	return &Folder{
		id:              id,
		name:            name,
		icon:            icon,
		workspaceID:     workspaceID,
		folderHierarchy: folderHierarchy,
	}
}

func UnmarshalFolder(
	id uuid.UUID,
	name string,
	icon *string,
	workspaceID uuid.UUID,
	folderHierarchy FolderHierarchy,
	trashedBy *TrashedBy,
	trashedAt *time.Time,
) *Folder {
	return &Folder{
		id:              id,
		name:            name,
		icon:            icon,
		workspaceID:     workspaceID,
		folderHierarchy: folderHierarchy,
		trashedBy:       trashedBy,
		trashedAt:       trashedAt,
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
}

func (f *Folder) Icon() *string {
	return f.icon
}

func (f *Folder) SetIcon(icon string) {
	f.icon = &icon
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
}

func (f *Folder) TrashedBy() *TrashedBy {
	return f.trashedBy
}

func (f *Folder) TrashedByString() *string {
	if f.trashedBy == nil {
		return nil
	}
	return new(f.trashedBy.String())
}

func (f *Folder) TrashedAt() *time.Time {
	return f.trashedAt
}

func (f *Folder) Trash(trashedBy TrashedBy) {
	f.trashedBy = &trashedBy
	f.trashedAt = new(time.Now())
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
