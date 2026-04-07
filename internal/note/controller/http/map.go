package http

import (
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func toNote(n app.Note) note.Note {
	var icon *string
	if n.Icon != "" {
		icon = &n.Icon
	}

	var tags *[]string
	if len(n.Tags) > 0 {
		tags = &n.Tags
	}

	var trashed *note.NoteTrashed
	if n.Trashed != nil {
		trashed = &note.NoteTrashed{
			TrashedBy: toTrashedBy(n.Trashed.TrashedBy),
			TrashedAt: n.Trashed.TrashedAt,
		}
	}

	return note.Note{
		Id:        &n.ID,
		Name:      n.Name,
		Icon:      icon,
		Tags:      tags,
		FolderId:  &n.FolderID,
		UpdatedAt: &n.UpdatedAt,
		Trashed:   trashed,
	}
}

func toFolder(f app.Folder) note.Folder {
	var icon *string
	if f.Icon != "" {
		icon = &f.Icon
	}

	var parentID *uuid.UUID
	if f.ParentID != uuid.Nil {
		parentID = &f.ParentID
	}

	var trashed *note.FolderTrashed
	if f.Trashed != nil {
		trashed = &note.FolderTrashed{
			TrashedBy: toTrashedBy(f.Trashed.TrashedBy),
			TrashedAt: f.Trashed.TrashedAt,
		}
	}

	return note.Folder{
		Id:          &f.ID,
		Name:        f.Name,
		Icon:        icon,
		ParentId:    parentID,
		WorkspaceId: &f.WorkspaceID,
		UpdatedAt:   &f.UpdatedAt,
		Trashed:     trashed,
	}
}

func toWorkspace(w app.Workspace) note.Workspace {
	return note.Workspace{
		Id:   &w.ID,
		Name: w.Name,
		Slug: w.Slug,
	}
}

func toWorkspaceRole(r app.WorkspaceRole) note.WorkspaceRole {
	switch r {
	case app.WorkspaceRoleOwner:
		return note.Owner
	case app.WorkspaceRoleEditor:
		return note.Editor
	case app.WorkspaceRoleViewer:
		return note.Viewer
	default:
		panic("invalid workspace role")
	}
}

func toWorkspaceMember(m *app.WorkspaceMember) note.WorkspaceMember {
	var username *string
	if m.Username != "" {
		username = &m.Username
	}

	return note.WorkspaceMember{
		Id:       m.ID,
		Role:     toWorkspaceRole(m.Role),
		Username: username,
	}
}

func toWorkspaceTreeNote(n *app.WorkspaceTreeNote) note.WorkspaceTreeNote {
	var icon *string
	if n.Icon != "" {
		icon = &n.Icon
	}
	return note.WorkspaceTreeNote{
		Id:        &n.ID,
		Name:      n.Name,
		Icon:      icon,
		UpdatedAt: &n.UpdatedAt,
	}
}

func toWorkspaceTreeFolder(f *app.WorkspaceTreeFolder) note.WorkspaceTreeFolder {
	var icon *string
	if f.Icon != "" {
		icon = &f.Icon
	}
	notes := make([]note.WorkspaceTreeNote, len(f.Notes))
	for i, n := range f.Notes {
		notes[i] = toWorkspaceTreeNote(n)
	}
	children := make([]note.WorkspaceTreeFolder, len(f.Children))
	for i, c := range f.Children {
		children[i] = toWorkspaceTreeFolder(c)
	}
	return note.WorkspaceTreeFolder{
		Id:        &f.ID,
		Name:      f.Name,
		Icon:      icon,
		Notes:     notes,
		Children:  children,
		UpdatedAt: &f.UpdatedAt,
	}
}

func toTrashedFolder(f *app.TrashedFolder) note.TrashedFolder {
	return note.TrashedFolder{
		Id:   f.ID,
		Name: &f.Name,
		Trashed: note.Trashed{
			TrashedBy: toTrashedBy(f.Trashed.TrashedBy),
			TrashedAt: f.Trashed.TrashedAt,
		},
	}
}

func toTrashedNote(n *app.TrashedNote) note.TrashedNote {
	return note.TrashedNote{
		Id:   n.ID,
		Name: &n.Name,
		Trashed: note.Trashed{
			TrashedBy: toTrashedBy(n.Trashed.TrashedBy),
			TrashedAt: n.Trashed.TrashedAt,
		},
	}
}

func toNoteLink(n *app.NoteLink) note.NoteLink {
	return note.NoteLink{
		Id:   &n.ID,
		Name: n.Name,
		Icon: &n.Icon,
	}
}

func toGraph(g *app.Graph) note.Graph {
	nodes := make([]note.GraphNode, len(g.Nodes))
	for i, n := range g.Nodes {
		nodes[i].Id = n.ID
		nodes[i].Name = n.Name
		nodes[i].Type = note.GraphNodesType(n.Type)
		if n.Weight != 0 {
			w := float32(n.Weight)
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
		panic("invalid trashed by")
	}
}
