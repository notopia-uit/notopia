package query

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

type NoteLink struct {
	Id   uuid.UUID
	Name string
	Icon *string
}

type NoteLinkResult struct {
	OutgoingLinks []NoteLink
	Backlinks     []NoteLink
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

type Trash struct {
	Notes   []TrashedNote
	Folders []TrashedFolder
}

type CheckWorkspaceExistsResult struct {
	Exists bool
}

type WorkspaceMembersUpdatedEvent struct {
	Id      uuid.UUID
	Members []WorkspaceMember
}
