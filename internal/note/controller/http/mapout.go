package http

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func toNoteDTO(n *app.Note) (note.Note, error) {
	var icon *string
	if n.Icon != "" {
		icon = &n.Icon
	}

	var tags *[]string
	if len(n.Tags) > 0 {
		tags = &n.Tags
	}

	var trashed *note.NoteTrashed
	if n.Trashed.IsTrashed() {
		trashedBy, err := toTrashedByDTO(n.Trashed.By)
		if err != nil {
			return note.Note{}, fmt.Errorf("invalid trashed by: %v", err)
		}
		trashed = &note.NoteTrashed{
			By: trashedBy,
			At: n.Trashed.At,
		}
	}

	return note.Note{
		Id:        &n.ID,
		Name:      n.Name,
		Icon:      icon,
		Tags:      tags,
		FolderId:  n.FolderID,
		UpdatedAt: &n.UpdatedAt,
		Trashed:   trashed,
	}, nil
}

func toFolderDTO(f *app.Folder) (note.Folder, error) {
	var icon *string
	if f.Icon != "" {
		icon = &f.Icon
	}

	var parentID *uuid.UUID
	if f.ParentID != uuid.Nil {
		parentID = &f.ParentID
	}

	var trashed *note.FolderTrashed
	if f.Trashed.IsTrashed() {
		trashedBy, err := toTrashedByDTO(f.Trashed.By)
		if err != nil {
			return note.Folder{}, fmt.Errorf("invalid trashed by: %v", err)
		}
		trashed = &note.FolderTrashed{
			By: trashedBy,
			At: f.Trashed.At,
		}
	}

	return note.Folder{
		Id:          &f.ID,
		Name:        f.Name,
		Icon:        icon,
		ParentId:    parentID,
		WorkspaceId: f.WorkspaceID,
		UpdatedAt:   &f.UpdatedAt,
		Trashed:     trashed,
	}, nil
}

func toWorkspaceDTO(w *app.Workspace) note.Workspace {
	return note.Workspace{
		Id:   &w.ID,
		Name: w.Name,
		Slug: w.Slug,
	}
}

func toWorkspaceRoleDTO(r app.WorkspaceRole) (note.WorkspaceRole, error) {
	switch r {
	case app.WorkspaceRoleOwner:
		return note.Owner, nil
	case app.WorkspaceRoleEditor:
		return note.Editor, nil
	case app.WorkspaceRoleViewer:
		return note.Viewer, nil
	case app.WorkspaceRoleUnspecified:
		return note.WorkspaceRole(""), errs.NewInternal("unspecified workspace role")
	default:
		return note.WorkspaceRole(""), errs.NewInternal(fmt.Sprintf("invalid workspace role: %v", r))
	}
}

func toUserWorkspaceDTO(u *app.UserWorkspace) (note.UserWorkspace, error) {
	role, err := toWorkspaceRoleDTO(u.Role)
	if err != nil {
		return note.UserWorkspace{}, err
	}
	return note.UserWorkspace{
		Workspace: toWorkspaceDTO(&u.Workspace),
		Role:      role,
	}, nil
}

func toWorkspaceMemberDTO(m *app.WorkspaceMember) (note.WorkspaceMember, error) {
	var name *string
	if m.Name != "" {
		name = &m.Name
	}
	role, err := toWorkspaceRoleDTO(m.Role)
	if err != nil {
		return note.WorkspaceMember{}, err
	}

	return note.WorkspaceMember{
		Id:   m.ID,
		Role: role,
		Name: name,
	}, nil
}

func toWorkspaceMembersDTO(members []app.WorkspaceMember) ([]note.WorkspaceMember, error) {
	out := make([]note.WorkspaceMember, len(members))
	for i := range members {
		member, err := toWorkspaceMemberDTO(&members[i])
		if err != nil {
			return nil, errs.NewInternal(fmt.Sprintf("invalid workspace member: %v", err))
		}
		out[i] = member
	}
	return out, nil
}

func toWorkspaceTreeNoteDTO(n *app.WorkspaceTreeNote) note.WorkspaceTreeNote {
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

func toWorkspaceTreeFolderDTO(f *app.WorkspaceTreeFolder) note.WorkspaceTreeFolder {
	var icon *string
	if f.Icon != "" {
		icon = &f.Icon
	}
	notes := make([]note.WorkspaceTreeNote, len(f.Notes))
	for i := range f.Notes {
		notes[i] = toWorkspaceTreeNoteDTO(&f.Notes[i])
	}
	children := make([]note.WorkspaceTreeFolder, len(f.Children))
	for i := range f.Children {
		children[i] = toWorkspaceTreeFolderDTO(&f.Children[i])
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

func toTrashedFolderDTO(f *app.TrashedFolder) (note.TrashedFolder, error) {
	trashedBy, err := toTrashedByDTO(f.Trashed.By)
	if err != nil {
		return note.TrashedFolder{}, err
	}
	return note.TrashedFolder{
		Id:   f.ID,
		Name: &f.Name,
		Trashed: note.Trashed{
			By: trashedBy,
			At: f.Trashed.At,
		},
	}, nil
}

func toTrashedNoteDTO(n *app.TrashedNote) (note.TrashedNote, error) {
	trashedBy, err := toTrashedByDTO(n.Trashed.By)
	if err != nil {
		return note.TrashedNote{}, err
	}
	return note.TrashedNote{
		Id:   n.ID,
		Name: &n.Name,
		Trashed: note.Trashed{
			By: trashedBy,
			At: n.Trashed.At,
		},
	}, nil
}

func toNoteLinkDTO(n *app.NoteLink) note.NoteLink {
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

func toGraphDTO(g *app.Graph) note.Graph {
	nodes := make([]note.GraphNode, len(g.Nodes))
	for i, n := range g.Nodes {
		nodes[i].Id = n.ID
		nodes[i].Name = n.Name
		nodes[i].Type = note.GraphNodesType(n.Type)
		if n.Weight != 0 {
			nodes[i].Weight = &n.Weight
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

func toShowTrashDTO(t *app.Trash) (note.ShowTrash200JSONResponse, error) {
	notes := make([]note.TrashedNote, len(t.Notes))
	for i := range t.Notes {
		trashedNote, err := toTrashedNoteDTO(&t.Notes[i])
		if err != nil {
			return note.ShowTrash200JSONResponse{}, fmt.Errorf("invalid trashed note: %v", err)
		}
		notes[i] = trashedNote
	}
	folders := make([]note.TrashedFolder, len(t.Folders))
	for i := range t.Folders {
		trashedFolder, err := toTrashedFolderDTO(&t.Folders[i])
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

func toGetNoteLinksDTO(r *app.NoteLinkResult) note.GetNoteLinks200JSONResponse {
	outgoing := make([]note.NoteLink, len(r.OutgoingLinks))
	for i := range r.OutgoingLinks {
		outgoing[i] = toNoteLinkDTO(&r.OutgoingLinks[i])
	}
	backlinks := make([]note.NoteLink, len(r.Backlinks))
	for i := range r.Backlinks {
		backlinks[i] = toNoteLinkDTO(&r.Backlinks[i])
	}
	return note.GetNoteLinks200JSONResponse{
		OutgoingLinks: &outgoing,
		Backlinks:     &backlinks,
	}
}

func toTrashedByDTO(t app.TrashedBy) (note.TrashedBy, error) {
	switch t {
	case app.TrashedByParent:
		return note.Parent, nil
	case app.TrashedByPurpose:
		return note.Purpose, nil
	case app.TrashedByUnspecified:
		return "", errs.NewInternal("unspecified trashed by")
	default:
		return "", errs.NewInternal(fmt.Sprintf("invalid trashed by: %v", t))
	}
}

func toSearchTokenDTO(t *app.SearchToken) note.SearchToken {
	return note.SearchToken{
		Token:     t.Token,
		ExpiresAt: t.ExpiresAt,
	}
}
