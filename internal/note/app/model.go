package app

import (
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
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

type Note struct {
	ID                 uuid.UUID
	Name               string
	Icon               *string
	Tags               []string
	Size               int32
	FolderID           uuid.UUID
	BacklinksCount     int
	OutgoingLinksCount int
	UpdatedAt          time.Time
}

type Folder struct {
	ID          uuid.UUID
	Name        string
	Icon        *string
	UpdatedAt   time.Time
	ParentID    uuid.UUID
	WorkspaceID uuid.UUID
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
	Weight *float64
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
	Icon *string
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
	ID       uuid.UUID
	Username *string
	Role     WorkspaceRole
}

type WorkspaceTreeNote struct {
	ID        uuid.UUID
	Name      string
	Icon      *string
	UpdatedAt time.Time
}

type WorkspaceTreeFolder struct {
	ID        uuid.UUID
	Name      string
	Icon      *string
	Notes     []*WorkspaceTreeNote
	Children  []*WorkspaceTreeFolder
	UpdatedAt time.Time
}

type TrashedFolder struct {
	ID        uuid.UUID
	Name      string
	TrashedBy domain.TrashedBy
	TrashedAt time.Time
}

type TrashedNote struct {
	ID        uuid.UUID
	Name      string
	TrashedBy domain.TrashedBy
	TrashedAt time.Time
}

type Trash struct {
	Notes   []*TrashedNote
	Folders []*TrashedFolder
}

type CheckWorkspaceSlugExistsResult struct {
	Exists bool
}

type WorkspaceMembersUpdatedEvent struct {
	ID      uuid.UUID
	Members []*WorkspaceMember
}
