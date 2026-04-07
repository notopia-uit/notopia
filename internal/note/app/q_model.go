package app

import (
	"time"

	"github.com/google/uuid"
)

type PaginationParams struct {
	Page  int
	Limit int
}

type Pagination struct {
	Page       int
	Limit      int
	Total      int
	TotalPages int
	hasNext    bool
	hasPrev    bool
}

type Paginated[T any] struct {
	Data       []T
	Pagination Pagination
}

type Trashed struct {
	TrashedBy TrashedBy
	TrashedAt time.Time
}

type Note struct {
	ID                 uuid.UUID
	Name               string
	Icon               string
	Tags               []string
	Size               int32
	FolderID           uuid.UUID
	BacklinksCount     int
	OutgoingLinksCount int
	Trashed            *Trashed
	UpdatedAt          time.Time
}

type Folder struct {
	ID          uuid.UUID
	Name        string
	Icon        string
	ParentID    uuid.UUID
	WorkspaceID uuid.UUID
	Trashed     *Trashed
	UpdatedAt   time.Time
}

type GraphNodeType string

var (
	GraphNodeTypeNote GraphNodeType = "note"
	GraphNodeTypeTag  GraphNodeType = "tag"
)

type GraphNode struct {
	ID     string
	Name   string
	Type   GraphNodeType
	Weight float64
}

type GraphLink struct {
	Source string
	Target string
}

type Graph struct {
	Nodes []*GraphNode
	Links []*GraphLink
}

type NoteLink struct {
	ID   uuid.UUID
	Name string
	Icon string
}

type NoteLinkResult struct {
	OutgoingLinks []*NoteLink
	Backlinks     []*NoteLink
}

type Workspace struct {
	ID   uuid.UUID
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
	ID       string
	Username string
	Role     WorkspaceRole
}

type WorkspaceTreeNote struct {
	ID        uuid.UUID
	Name      string
	Icon      string
	UpdatedAt time.Time
}

type WorkspaceTreeFolder struct {
	ID        uuid.UUID
	Name      string
	Icon      string
	Notes     []*WorkspaceTreeNote
	Children  []*WorkspaceTreeFolder
	UpdatedAt time.Time
}

type TrashedBy string

const (
	TrashedByPurpose TrashedBy = "purpose"
	TrashedByParent  TrashedBy = "parent"
)

type TrashedFolder struct {
	ID      uuid.UUID
	Name    string
	Icon    string
	Trashed Trashed
}

type TrashedNote struct {
	ID      uuid.UUID
	Name    string
	Icon    string
	Trashed Trashed
}

type Trash struct {
	Notes   []*TrashedNote
	Folders []*TrashedFolder
}

type CheckWorkspaceSlugExistsResult struct {
	Exists bool
}
