package http

import (
	"github.com/notopia-uit/notopia/internal/note/app/query"
	"github.com/notopia-uit/notopia/internal/note/domain"
	note "github.com/notopia-uit/notopia/pkg/api/note"
)

func workspaceEventToDTO(event domain.WorkspaceEvent) (any, bool) {
	var dto any
	switch e := event.(type) {
	case *domain.FolderCreatedEvent:
		dto = &note.FolderCreatedEvent{
			Data: struct {
				Id   *note.Id  `json:"id,omitempty"`
				Name note.Name `json:"name"`
			}{
				Id:   &e.Id,
				Name: e.Name,
			},
			Type: note.FolderCreatedEventTypeFolderCreatedEvent,
		}
	case *domain.FolderDeletedEvent:
		dto = &note.FolderDeletedEvent{
			Data: struct {
				Id *note.Id `json:"id,omitempty"`
			}{
				Id: &e.Id,
			},
			Type: note.FolderDeletedEventTypeFolderDeletedEvent,
		}
	case *domain.FolderUpdatedEvent:
		folder := (*domain.Folder)(e)
		id := folder.ID()
		workspaceID := folder.WorkspaceID()
		dto = &note.FolderUpdatedEvent{
			Data: note.Folder{
				Id:          &id,
				Name:        folder.Name(),
				Icon:        folder.Icon(),
				ParentId:    folder.ParentID(),
				WorkspaceId: &workspaceID,
			},
			Type: note.FolderInfoUpdatedEvent,
		}
	case *domain.NoteCreatedEvent:
		dto = &note.NoteCreatedEvent{
			Data: struct {
				Id   *note.NotePropertiesId `json:"id,omitempty"`
				Name note.PropertiesName    `json:"name"`
			}{
				Id:   &e.Id,
				Name: e.Name,
			},
			Type: note.NoteCreatedEventTypeNoteCreatedEvent,
		}
	case *domain.NoteDeletedEvent:
		dto = &note.NoteDeletedEvent{
			Data: struct {
				Id *note.NotePropertiesId `json:"id,omitempty"`
			}{
				Id: &e.Id,
			},
			Type: note.NoteDeletedEventTypeNoteDeletedEvent,
		}
	case *domain.NoteUpdatedEvent:
		n := (*domain.Note)(e)
		dto = &note.NoteUpdatedEvent{
			Data: note.Note{
				Id:       new(n.ID()),
				Name:     n.Name(),
				Icon:     n.Icon(),
				Tags:     n.Tags(),
				FolderId: new(n.FolderID()),
			},
			Type: note.NoteUpdatedEventTypeNoteUpdatedEvent,
		}
	case *domain.WorkspaceUpdatedEvent:
		w := (*domain.Workspace)(e)
		dto = &note.WorkspaceUpdatedEvent{
			Data: note.Workspace{
				Id:   new(w.ID()),
				Name: w.Name(),
				Slug: w.Slug(),
			},
			Type: note.WorkspaceUpdatedEventTypeWorkspaceUpdatedEvent,
		}
	}
	return dto, dto != nil
}

func queryNoteToDTO(n query.Note) note.Note {
	id := n.Id
	folderID := n.FolderId
	backlinksCount := n.BacklinksCount
	outgoingLinksCount := n.OutgoingLinksCount
	updatedAt := n.UpdatedAt
	tags := n.Tags
	if tags == nil {
		tags = []string{}
	}
	return note.Note{
		Id:                 &id,
		Name:               n.Name,
		Icon:               n.Icon,
		Tags:               tags,
		FolderId:           &folderID,
		BacklinksCount:     &backlinksCount,
		OutgoingLinksCount: &outgoingLinksCount,
		UpdatedAt:          &updatedAt,
	}
}

func queryFolderToDTO(f query.Folder) note.Folder {
	id := f.Id
	parentID := f.ParentId
	workspaceID := f.WorkspaceId
	updatedAt := f.UpdatedAt
	return note.Folder{
		Id:          &id,
		Name:        f.Name,
		Icon:        f.Icon,
		ParentId:    &parentID,
		WorkspaceId: &workspaceID,
		UpdatedAt:   &updatedAt,
	}
}

func queryWorkspaceToDTO(w query.Workspace) note.Workspace {
	id := w.Id
	return note.Workspace{
		Id:   &id,
		Name: w.Name,
		Slug: w.Slug,
	}
}

func queryWorkspaceMemberToDTO(m query.WorkspaceMember) note.WorkspaceMember {
	return note.WorkspaceMember{
		Role:     note.WorkspaceRole(m.Role),
		Username: m.Username,
	}
}

func queryWorkspaceTreeNoteToDTO(n query.WorkspaceTreeNote) note.WorkspaceTreeNote {
	id := n.Id
	updatedAt := n.UpdatedAt
	return note.WorkspaceTreeNote{
		Id:        &id,
		Name:      n.Name,
		Icon:      n.Icon,
		UpdatedAt: &updatedAt,
	}
}

func queryWorkspaceTreeFolderToDTO(f query.WorkspaceTreeFolder) note.WorkspaceTreeFolder {
	id := f.Id
	updatedAt := f.UpdatedAt
	notes := make([]note.WorkspaceTreeNote, len(f.Notes))
	for i, n := range f.Notes {
		notes[i] = queryWorkspaceTreeNoteToDTO(n)
	}
	children := make([]note.WorkspaceTreeFolder, len(f.Children))
	for i, c := range f.Children {
		children[i] = queryWorkspaceTreeFolderToDTO(c)
	}
	return note.WorkspaceTreeFolder{
		Id:        &id,
		Name:      f.Name,
		Icon:      f.Icon,
		Notes:     notes,
		Children:  children,
		UpdatedAt: &updatedAt,
	}
}

func queryTrashedFolderToDTO(f query.TrashedFolder) note.TrashedFolder {
	name := f.Name
	trashedAt := f.TrashedAt
	trashedBy := note.TrashedBy(f.TrashedBy)
	return note.TrashedFolder{
		Id:        f.Id,
		Name:      &name,
		TrashedAt: &trashedAt,
		TrashedBy: &trashedBy,
	}
}

func queryTrashedNoteToDTO(n query.TrashedNote) note.TrashedNote {
	name := n.Name
	trashedAt := n.TrashedAt
	trashedBy := note.TrashedBy(n.TrashedBy)
	return note.TrashedNote{
		Id:        n.Id,
		Name:      &name,
		TrashedAt: &trashedAt,
		TrashedBy: &trashedBy,
	}
}

func queryNoteLinkToDTO(n query.NoteLink) note.NoteLink {
	id := n.Id
	return note.NoteLink{
		Id:   &id,
		Name: n.Name,
		Icon: n.Icon,
	}
}

func queryGraphToDTO(g query.Graph) note.Graph {
	dto := note.Graph{
		Nodes: make([]struct {
			Id     string              `json:"id"`
			Name   string              `json:"name"`
			Type   note.GraphNodesType `json:"type"`
			Weight *float32            `json:"weight,omitempty"`
		}, len(g.Nodes)),
		Links: make([]struct {
			Source string `json:"source"`
			Target string `json:"target"`
		}, len(g.Links)),
	}
	for i, n := range g.Nodes {
		dto.Nodes[i].Id = n.Id
		dto.Nodes[i].Name = n.Name
		dto.Nodes[i].Type = note.GraphNodesType(n.Type)
		if n.Weight != nil {
			w := float32(*n.Weight)
			dto.Nodes[i].Weight = &w
		}
	}
	for i, l := range g.Links {
		dto.Links[i].Source = l.Source
		dto.Links[i].Target = l.Target
	}
	return dto
}

func queryWorkspaceMembersUpdatedEventToDTO(e query.WorkspaceMembersUpdatedEvent) note.WorkspaceMemebersUpdatedEvent {
	id := e.Id
	members := make([]note.WorkspaceMember, len(e.Members))
	for i, m := range e.Members {
		members[i] = queryWorkspaceMemberToDTO(m)
	}
	return note.WorkspaceMemebersUpdatedEvent{
		Data: struct {
			Id      *note.PropertiesId      `json:"id,omitempty"`
			Members *[]note.WorkspaceMember `json:"members,omitempty"`
		}{
			Id:      &id,
			Members: &members,
		},
		Type: note.WorkspaceMembersUpdatedEvent,
	}
}
