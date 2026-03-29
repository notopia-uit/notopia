package http

import (
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func toNote(n app.Note) note.Note {
	id := n.ID
	folderID := n.FolderID
	updatedAt := n.UpdatedAt
	tags := n.Tags
	if tags == nil {
		tags = []string{}
	}
	return note.Note{
		Id:        &id,
		Name:      n.Name,
		Icon:      n.Icon,
		Tags:      tags,
		FolderId:  &folderID,
		UpdatedAt: &updatedAt,
	}
}

func toFolder(f app.Folder) note.Folder {
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

func toWorkspace(w app.Workspace) note.Workspace {
	id := w.ID
	return note.Workspace{
		Id:   &id,
		Name: w.Name,
		Slug: w.Slug,
	}
}

func toWorkspaceMember(m *app.WorkspaceMember) note.WorkspaceMember {
	return note.WorkspaceMember{
		Id:       m.ID,
		Role:     note.WorkspaceRole(m.Role),
		Username: m.Username,
	}
}

func toWorkspaceTreeNote(n *app.WorkspaceTreeNote) note.WorkspaceTreeNote {
	id := n.ID
	updatedAt := n.UpdatedAt
	return note.WorkspaceTreeNote{
		Id:        &id,
		Name:      n.Name,
		Icon:      n.Icon,
		UpdatedAt: &updatedAt,
	}
}

func toWorkspaceTreeFolder(f *app.WorkspaceTreeFolder) note.WorkspaceTreeFolder {
	id := f.ID
	updatedAt := f.UpdatedAt
	notes := make([]note.WorkspaceTreeNote, len(f.Notes))
	for i, n := range f.Notes {
		notes[i] = toWorkspaceTreeNote(n)
	}
	children := make([]note.WorkspaceTreeFolder, len(f.Children))
	for i, c := range f.Children {
		children[i] = toWorkspaceTreeFolder(c)
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

func toTrashedFolder(f *app.TrashedFolder) note.TrashedFolder {
	name := f.Name
	trashedAt := f.TrashedAt
	return note.TrashedFolder{
		Id:        f.ID,
		Name:      &name,
		TrashedAt: &trashedAt,
		TrashedBy: toTrashedBy(f.TrashedBy),
	}
}

func toTrashedNote(n *app.TrashedNote) note.TrashedNote {
	name := n.Name
	trashedAt := n.TrashedAt
	return note.TrashedNote{
		Id:        n.ID,
		Name:      &name,
		TrashedAt: &trashedAt,
		TrashedBy: toTrashedBy(n.TrashedBy),
	}
}

func toNoteLink(n *app.NoteLink) note.NoteLink {
	id := n.ID
	return note.NoteLink{
		Id:   &id,
		Name: n.Name,
		Icon: n.Icon,
	}
}

func toGraph(g *app.Graph) note.Graph {
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

func toShowTrash(t *app.Trash) note.ShowTrash200JSONResponse {
	notes := make([]note.TrashedNote, len(t.Notes))
	for i, n := range t.Notes {
		notes[i] = toTrashedNote(n)
	}
	folders := make([]note.TrashedFolder, len(t.Folders))
	for i, f := range t.Folders {
		folders[i] = toTrashedFolder(f)
	}
	return note.ShowTrash200JSONResponse{
		Notes:   notes,
		Folders: folders,
	}
}

func toGetNoteLinks(r *app.NoteLinkResult) note.GetNoteLinks200JSONResponse {
	outgoing := make([]note.NoteLink, len(r.OutgoingLinks))
	for i, l := range r.OutgoingLinks {
		outgoing[i] = toNoteLink(l)
	}
	backlinks := make([]note.NoteLink, len(r.Backlinks))
	for i, l := range r.Backlinks {
		backlinks[i] = toNoteLink(l)
	}
	return note.GetNoteLinks200JSONResponse{
		OutgoingLinks: &outgoing,
		Backlinks:     &backlinks,
	}
}

func toTrashedBy(t app.TrashedBy) note.TrashedBy {
	switch t {
	case app.TrashedByParent:
		return note.Parent
	case app.TrashedByPurpose:
		return note.Purpose
	default:
		panic("unknown trashed by")
	}
}
