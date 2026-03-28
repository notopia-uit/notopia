package http

import (
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func workspaceEventToDTO(event domain.Event) (any, bool) {
	var dto any
	switch e := event.(type) {
	case *domain.FolderCreatedEvent:
		dto = &note.FolderCreatedEvent{
			Data: note.FolderCreatedEventData{
				Icon: e.Icon,
				Id:   &e.AggregateID,
				Name: e.Name,
			},
			Type: note.FolderCreatedEventTypeFolderCreatedEvent,
		}
	case *domain.FolderDeletedEvent:
		dto = &note.FolderDeletedEvent{
			Data: note.FolderDeletedEventData{
				Id: &e.AggregateID,
			},
			Type: note.FolderDeletedEventTypeFolderDeletedEvent,
		}
	case *domain.FolderUpdatedEvent:
		dto = &note.FolderUpdatedEvent{
			Data: note.FolderUpdatedEventData{
				Id:   &e.AggregateID,
				Name: e.Name,
				Icon: e.Icon,
			},
			Type: note.FolderUpdatedEventTypeFolderUpdatedEvent,
		}
	case *domain.FolderMovedEvent:
		dto = &note.FolderMovedEvent{
			Data: note.FolderMovedEventData{
				Id:       &e.AggregateID,
				ParentId: &e.ParentID,
			},
			Type: note.FolderMovedEventTypeFolderMovedEvent,
		}
	case *domain.FolderTrashedEvent:
		dto = &note.FolderTrashedEvent{
			Data: note.FolderTrashedEventData{
				Id: &e.AggregateID,
			},
			Type: note.FolderTrashedEventTypeFolderTrashedEvent,
		}
	case *domain.FolderRestoredEvent:
		dto = &note.FolderRestoredEvent{
			Data: note.FolderRestoredEventData{
				Id: &e.AggregateID,
			},
			Type: note.FolderRestoredEventTypeFolderRestoredEvent,
		}
	case *domain.FolderPermanentlyDeletedEvent:
		dto = &note.FolderPermanentlyDeletedEvent{
			Data: note.FolderPermanentlyDeletedEventData{
				Id: &e.AggregateID,
			},
			Type: note.FolderPermanentlyDeletedEventTypeFolderPermanentlyDeletedEvent,
		}
	case *domain.NoteCreatedEvent:
		dto = &note.NoteCreatedEvent{
			Data: note.NoteCreatedEventData{
				Id:   &e.AggregateID,
				Name: e.Name,
				Icon: e.Icon,
			},
			Type: note.NoteCreatedEventTypeNoteCreatedEvent,
		}
	case *domain.NoteDeletedEvent:
		dto = &note.NoteDeletedEvent{
			Data: note.NoteDeletedEventData{
				Id: &e.AggregateID,
			},
			Type: note.NoteDeletedEventTypeNoteDeletedEvent,
		}
	case *domain.NoteUpdatedEvent:
		dto = &note.NoteUpdatedEvent{
			Data: note.Note{
				Id:       &e.AggregateID,
				Name:     e.Name,
				Icon:     e.Icon,
				Tags:     e.Tags,
				FolderId: &e.FolderID,
			},
			Type: note.NoteUpdatedEventTypeNoteUpdatedEvent,
		}
	case *domain.NoteMovedEvent:
		dto = &note.NoteMovedEvent{
			Data: note.NoteMovedEventData{
				Id:       &e.AggregateID,
				FolderId: &e.FolderID,
			},
			Type: note.NoteMovedEventTypeNoteMovedEvent,
		}
	case *domain.NoteTrashedEvent:
		dto = &note.NoteTrashedEvent{
			Data: note.NoteTrashedEventData{
				Id: &e.AggregateID,
			},
			Type: note.NoteTrashedEventTypeNoteTrashedEvent,
		}
	case *domain.NoteRestoredEvent:
		dto = &note.NoteRestoredEvent{
			Data: note.NoteRestoredEventData{
				Id: &e.AggregateID,
			},
			Type: note.NoteRestoredEventTypeNoteRestoredEvent,
		}
	case *domain.NotePermanentlyDeletedEvent:
		dto = &note.NotePermanentlyDeletedEvent{
			Data: note.NotePermanentlyDeletedEventData{
				Id: &e.AggregateID,
			},
			Type: note.NotePermanentlyDeletedEventTypeNotePermanentlyDeletedEvent,
		}
	case *domain.WorkspaceUpdatedEvent:
		dto = &note.WorkspaceUpdatedEvent{
			Data: note.Workspace{
				Id:   &e.AggregateID,
				Name: e.Name,
				Slug: e.Slug,
			},
			Type: note.WorkspaceUpdatedEventTypeWorkspaceUpdatedEvent,
		}
	case *domain.WorkspaceDeletedEvent:
		dto = &note.WorkspaceDeletedEvent{
			Data: note.WorkspaceDeletedEventData{
				Id: &e.AggregateID,
			},
			Type: note.WorkspaceDeletedEventTypeWorkspaceDeletedEvent,
		}
	}
	return dto, dto != nil
}

func getNoteToDTO(n app.Note) note.Note {
	id := n.ID
	folderID := n.FolderID
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

func getFolderToDTO(f app.Folder) note.Folder {
	id := f.ID
	parentID := f.ParentID
	workspaceID := f.WorkspaceID
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

func getWorkspaceToDTO(w app.Workspace) note.Workspace {
	id := w.ID
	return note.Workspace{
		Id:   &id,
		Name: w.Name,
		Slug: w.Slug,
	}
}

func getWorkspaceMemberToDTO(m *app.WorkspaceMember) note.WorkspaceMember {
	return note.WorkspaceMember{
		Id:       m.ID,
		Role:     note.WorkspaceRole(m.Role),
		Username: m.Username,
	}
}

func getWorkspaceTreeNoteToDTO(n *app.WorkspaceTreeNote) note.WorkspaceTreeNote {
	id := n.ID
	updatedAt := n.UpdatedAt
	return note.WorkspaceTreeNote{
		Id:        &id,
		Name:      n.Name,
		Icon:      n.Icon,
		UpdatedAt: &updatedAt,
	}
}

func getWorkspaceTreeFolderToDTO(f *app.WorkspaceTreeFolder) note.WorkspaceTreeFolder {
	id := f.ID
	updatedAt := f.UpdatedAt
	notes := make([]note.WorkspaceTreeNote, len(f.Notes))
	for i, n := range f.Notes {
		notes[i] = getWorkspaceTreeNoteToDTO(n)
	}
	children := make([]note.WorkspaceTreeFolder, len(f.Children))
	for i, c := range f.Children {
		children[i] = getWorkspaceTreeFolderToDTO(c)
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

func getTrashedFolderToDTO(f *app.TrashedFolder) note.TrashedFolder {
	name := f.Name
	trashedAt := f.TrashedAt
	trashedBy := note.TrashedBy(f.TrashedBy)
	return note.TrashedFolder{
		Id:        f.ID,
		Name:      &name,
		TrashedAt: &trashedAt,
		TrashedBy: &trashedBy,
	}
}

func getTrashedNoteToDTO(n *app.TrashedNote) note.TrashedNote {
	name := n.Name
	trashedAt := n.TrashedAt
	trashedBy := note.TrashedBy(n.TrashedBy)
	return note.TrashedNote{
		Id:        n.ID,
		Name:      &name,
		TrashedAt: &trashedAt,
		TrashedBy: &trashedBy,
	}
}

func getNoteLinkToDTO(n *app.NoteLink) note.NoteLink {
	id := n.ID
	return note.NoteLink{
		Id:   &id,
		Name: n.Name,
		Icon: n.Icon,
	}
}

func getGraphToDTO(g *app.Graph) note.Graph {
	nodes := make([]note.GraphNode, len(g.Nodes))
	for i, n := range g.Nodes {
		nodes[i].Id = n.ID
		nodes[i].Name = n.Name
		nodes[i].Type = note.GraphNodesType(n.Type)
		if n.Weight != nil {
			w := float32(*n.Weight)
			nodes[i].Weight = &w
		}
	}
	links := make([]note.GraphLink, len(g.Links))
	for i, l := range g.Links {
		links[i].Source = l.Source
		links[i].Target = l.Target
	}
	return note.Graph{
		Nodes: nodes,
		Links: links,
	}
}

func getWorkspaceMembersUpdatedEventToDTO(e *app.WorkspaceMembersUpdatedEvent) note.WorkspaceMemebersUpdatedEvent {
	id := e.ID
	members := make([]note.WorkspaceMember, len(e.Members))
	for i, m := range e.Members {
		members[i] = getWorkspaceMemberToDTO(m)
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

func getTrashedToDTO(t *app.Trash) note.ShowTrash200JSONResponse {
	notes := make([]note.TrashedNote, len(t.Notes))
	for i, n := range t.Notes {
		notes[i] = getTrashedNoteToDTO(n)
	}
	folders := make([]note.TrashedFolder, len(t.Folders))
	for i, f := range t.Folders {
		folders[i] = getTrashedFolderToDTO(f)
	}
	return note.ShowTrash200JSONResponse{
		Notes:   notes,
		Folders: folders,
	}
}

func getNoteLinkResultToDTO(r *app.NoteLinkResult) note.GetNoteLinks200JSONResponse {
	outgoing := make([]note.NoteLink, len(r.OutgoingLinks))
	for i, l := range r.OutgoingLinks {
		outgoing[i] = getNoteLinkToDTO(l)
	}
	backlinks := make([]note.NoteLink, len(r.Backlinks))
	for i, l := range r.Backlinks {
		backlinks[i] = getNoteLinkToDTO(l)
	}
	return note.GetNoteLinks200JSONResponse{
		OutgoingLinks: &outgoing,
		Backlinks:     &backlinks,
	}
}
