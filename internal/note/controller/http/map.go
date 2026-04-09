package http

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func toNote(n app.Note) (note.Note, error) {
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
		trashedBy, err := toTrashedBy(n.Trashed.TrashedBy)
		if err != nil {
			return note.Note{}, fmt.Errorf("invalid trashed by: %v", err)
		}
		trashed = &note.NoteTrashed{
			TrashedBy: trashedBy,
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
	}, nil
}

func toFolder(f app.Folder) (note.Folder, error) {
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
		trashedBy, err := toTrashedBy(f.Trashed.TrashedBy)
		if err != nil {
			return note.Folder{}, fmt.Errorf("invalid trashed by: %v", err)
		}
		trashed = &note.FolderTrashed{
			TrashedBy: trashedBy,
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
	}, nil
}

func toWorkspace(w app.Workspace) note.Workspace {
	return note.Workspace{
		Id:   &w.ID,
		Name: w.Name,
		Slug: w.Slug,
	}
}

func toWorkspaceRole(r app.WorkspaceRole) (note.WorkspaceRole, error) {
	switch r {
	case app.WorkspaceRoleOwner:
		return note.Owner, nil
	case app.WorkspaceRoleEditor:
		return note.Editor, nil
	case app.WorkspaceRoleViewer:
		return note.Viewer, nil
	case app.WorkspaceRoleUnspecified:
		return note.WorkspaceRole(""), errs.NewInternal("unspecified workspace role", nil)
	default:
		return note.WorkspaceRole(""), errs.NewInternal(fmt.Sprintf("invalid workspace role: %v", r), nil)
	}
}

func toWorkspaceMember(m *app.WorkspaceMember) (note.WorkspaceMember, error) {
	var username *string
	if m.Username != "" {
		username = &m.Username
	}
	role, err := toWorkspaceRole(m.Role)
	if err != nil {
		return note.WorkspaceMember{}, err
	}

	return note.WorkspaceMember{
		Id:       m.ID,
		Role:     role,
		Username: username,
	}, nil
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

func toTrashedFolder(f *app.TrashedFolder) (note.TrashedFolder, error) {
	trashedBy, err := toTrashedBy(f.Trashed.TrashedBy)
	if err != nil {
		return note.TrashedFolder{}, err
	}
	return note.TrashedFolder{
		Id:   f.ID,
		Name: &f.Name,
		Trashed: note.Trashed{
			TrashedBy: trashedBy,
			TrashedAt: f.Trashed.TrashedAt,
		},
	}, nil
}

func toTrashedNote(n *app.TrashedNote) (note.TrashedNote, error) {
	trashedBy, err := toTrashedBy(n.Trashed.TrashedBy)
	if err != nil {
		return note.TrashedNote{}, err
	}
	return note.TrashedNote{
		Id:   n.ID,
		Name: &n.Name,
		Trashed: note.Trashed{
			TrashedBy: trashedBy,
			TrashedAt: n.Trashed.TrashedAt,
		},
	}, nil
}

func toNoteLink(n *app.NoteLink) note.NoteLink {
	var icon *string
	if n.Icon != "" {
		icon = &n.Icon
	}
	return note.NoteLink{
		Id:   &n.ID,
		Name: n.Name,
		Icon: icon,
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

func toShowTrash(t *app.Trash) (note.ShowTrash200JSONResponse, error) {
	notes := make([]note.TrashedNote, len(t.Notes))
	for i, n := range t.Notes {
		trashedNote, err := toTrashedNote(n)
		if err != nil {
			return note.ShowTrash200JSONResponse{}, fmt.Errorf("invalid trashed note: %v", err)
		}
		notes[i] = trashedNote
	}
	folders := make([]note.TrashedFolder, len(t.Folders))
	for i, f := range t.Folders {
		trashedFolder, err := toTrashedFolder(f)
		if err != nil {
			return note.ShowTrash200JSONResponse{}, fmt.Errorf("invalid trashed folder: %v", err)
		}
		folders[i] = trashedFolder
	}
	return note.ShowTrash200JSONResponse{
		Notes:   notes,
		Folders: folders,
	}, nil
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

func toTrashedBy(t app.TrashedBy) (note.TrashedBy, error) {
	switch t {
	case app.TrashedByParent:
		return note.Parent, nil
	case app.TrashedByPurpose:
		return note.Purpose, nil
	case app.TrashedByUnspecified:
		return note.TrashedBy(""), errs.NewInternal("unspecified trashed by", nil)
	default:
		return note.TrashedBy(""), errs.NewInternal(fmt.Sprintf("invalid trashed by: %v", t), nil)
	}
}
