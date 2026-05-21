package app

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type SortOrder uint8

const (
	SortOrderUnspecified SortOrder = iota
	SortOrderAsc
	SortOrderDesc
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
	By TrashedBy
	At time.Time
}

func (trashed Trashed) IsTrashed() bool {
	return trashed.By != TrashedByUnspecified && !trashed.At.IsZero()
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
	Trashed            Trashed
	UpdatedAt          time.Time
}

type Folder struct {
	ID          uuid.UUID
	Name        string
	Icon        string
	ParentID    uuid.UUID
	WorkspaceID uuid.UUID
	Trashed     Trashed
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
	Weight float32
}

type GraphLink struct {
	Source string
	Target string
}

type Graph struct {
	Nodes []GraphNode
	Links []GraphLink
}

var _ slog.LogValuer = Graph{}

func (g Graph) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("nodes_count", len(g.Nodes)),
		slog.Int("links_count", len(g.Links)),
	)
}

type NoteLink struct {
	ID   uuid.UUID
	Name string
	Icon string
}

type NoteLinkResult struct {
	OutgoingLinks []NoteLink
	Backlinks     []NoteLink
}

type Workspace struct {
	ID   uuid.UUID
	Slug string
	Name string
}

type WorkspaceRole uint8

const (
	WorkspaceRoleUnspecified WorkspaceRole = iota
	WorkspaceRoleOwner
	WorkspaceRoleEditor
	WorkspaceRoleViewer
)

func (r WorkspaceRole) String() string {
	switch r {
	case WorkspaceRoleOwner:
		return "owner"
	case WorkspaceRoleEditor:
		return "editor"
	case WorkspaceRoleViewer:
		return "viewer"
	case WorkspaceRoleUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

type WorkspaceMember struct {
	ID   string
	Name string
	Role WorkspaceRole
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
	Notes     []WorkspaceTreeNote
	Children  []WorkspaceTreeFolder
	UpdatedAt time.Time
}

var _ slog.LogValuer = WorkspaceTreeFolder{}

func (f WorkspaceTreeFolder) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", f.ID.String()),
		slog.String("name", f.Name),
		slog.String("icon", f.Icon),
		slog.Int("notes_count", len(f.Notes)),
		slog.Int("direct_children_count", len(f.Children)),
		slog.Time("updated_at", f.UpdatedAt),
	)
}

type TrashedBy uint8

const (
	TrashedByUnspecified TrashedBy = iota
	TrashedByPurpose
	TrashedByParent
)

func (t TrashedBy) String() string {
	switch t {
	case TrashedByPurpose:
		return "purpose"
	case TrashedByParent:
		return "parent"
	case TrashedByUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

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
	Notes   []TrashedNote
	Folders []TrashedFolder
}

type WorkspaceMemberUpdate struct {
	ID   string // User ID
	Role WorkspaceRole
}

type UserWorkspace struct {
	Workspace Workspace
	Role      WorkspaceRole
}

type User struct {
	ID     string
	Name   string
	Email  string // Can be empty
	Groups []string
	Roles  []string
}

type SearchToken struct {
	Token     string
	ExpiresAt time.Time
}
