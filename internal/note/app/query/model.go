package query

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type Note struct {
	Id                 uuid.UUID
	Name               string
	Icon               *string
	Tags               []string
	FolderId           uuid.UUID
	BacklinksCount     int
	OutgoingLinksCount int
	UpdatedAt          time.Time
}

type Folder struct {
	Id          uuid.UUID
	Name        string
	Icon        *string
	UpdatedAt   time.Time
	ParentId    uuid.UUID
	WorkspaceId uuid.UUID
}

type FolderCreatedEventData struct {
	Id   uuid.UUID
	Name string
}

type FolderCreatedEvent struct {
	Type string
	Data FolderCreatedEventData
}

type FolderDeletedEventData struct {
	Id uuid.UUID
}

type FolderDeletedEvent struct {
	Type string
	Data FolderDeletedEventData
}

type FolderUpdatedEvent struct {
	Type string
	Data Folder
}

type GraphNode struct {
	Id     string
	Name   string
	Type   string
	Weight *float64
}

type GraphLink struct {
	Source string
	Target string
}

type Graph struct {
	Nodes []GraphNode
	Links []GraphLink
}

type NoteCreatedEventData struct {
	Id   uuid.UUID
	Name string
}

type NoteCreatedEvent struct {
	Type string
	Data NoteCreatedEventData
}

type NoteDeletedEventData struct {
	Id uuid.UUID
}

type NoteDeletedEvent struct {
	Type string
	Data NoteDeletedEventData
}

type NoteLink struct {
	Id   uuid.UUID
	Name string
	Icon *string
}

type NoteUpdatedEvent struct {
	Type string
	Data Note
}

type TrashedFolder struct {
	Id        uuid.UUID
	Name      string
	TrashedBy domain.TrashedBy
	TrashedAt time.Time
}

type TrashedNote struct {
	Id        uuid.UUID
	Name      string
	TrashedBy domain.TrashedBy
	TrashedAt time.Time
}

type Workspace struct {
	Id   uuid.UUID
	Slug string
	Name string
}

type WorkspaceRole string

var (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleEditor WorkspaceRole = "editor"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

type WorkspaceMember struct {
	Id       uuid.UUID
	Username *string
	Role     WorkspaceRole
}

type WorkspaceMembersUpdatedEventData struct {
	Id      uuid.UUID
	Members []WorkspaceMember
}

type WorkspaceMembersUpdatedEvent struct {
	Type string
	Data WorkspaceMembersUpdatedEventData
}

type WorkspaceTreeNote struct {
	Id        uuid.UUID
	Name      string
	Icon      *string
	UpdatedAt time.Time
}

type WorkspaceTreeFolder struct {
	Id        uuid.UUID
	Name      string
	Icon      *string
	Notes     []WorkspaceTreeNote
	Children  []WorkspaceTreeFolder
	UpdatedAt time.Time
}

type WorkspaceUpdatedEvent struct {
	Type string
	Data Workspace
}
