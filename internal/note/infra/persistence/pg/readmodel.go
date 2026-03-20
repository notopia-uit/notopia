package pg

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type ReadModel struct {
	queries *pgsqlc.Queries
}

func NewReadModel(queries *pgsqlc.Queries) *ReadModel {
	return &ReadModel{queries: queries}
}

var ProvideReadModel = NewReadModel

var (
	_ app.GetWorkspaceTreeReadModel         = (*ReadModel)(nil)
	_ app.ShowTrashReadModel                = (*ReadModel)(nil)
	_ app.GetNoteGraphReadModel             = (*ReadModel)(nil)
	_ app.GetNoteLinksReadModel             = (*ReadModel)(nil)
	_ app.GetWorkspaceBySlugReadModel       = (*ReadModel)(nil)
	_ app.GetWorkspaceGraphReadModel        = (*ReadModel)(nil)
	_ app.CheckWorkspaceSlugExistsReadModel = (*ReadModel)(nil)
)

func (r *ReadModel) GetWorkspaceTree(ctx context.Context, q *app.GetWorkspaceTree) (*app.WorkspaceTreeFolder, error) {
	workspace, err := r.queries.GetWorkspace(ctx, &pgsqlc.GetWorkspaceParams{
		ID: &q.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceByIDNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	rootFolder, err := r.queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
		WorkspaceID:  &workspace.ID,
		IsRootFolder: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceRootFolderNotFound(workspace.ID, err)
		}
		return nil, toDomainError(err)
	}

	allFolders, err := r.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &workspace.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	allNotes, err := r.queries.GetNotesInWorkspace(ctx, workspace.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	folderMap := make(map[uuid.UUID]*pgsqlc.Folder)
	for _, folder := range allFolders {
		folderMap[folder.ID] = folder
	}

	notesByFolder := make(map[uuid.UUID][]*pgsqlc.Note)
	for _, note := range allNotes {
		notesByFolder[note.FolderID] = append(notesByFolder[note.FolderID], note)
	}

	tree := r.buildFolderTreeFromMap(*rootFolder, folderMap, notesByFolder)
	return tree, nil
}

func (r *ReadModel) buildFolderTreeFromMap(folder pgsqlc.Folder, folderMap map[uuid.UUID]*pgsqlc.Folder, notesByFolder map[uuid.UUID][]*pgsqlc.Note) *app.WorkspaceTreeFolder {
	result := app.WorkspaceTreeFolder{
		Id:        folder.ID,
		Name:      folder.Name,
		Icon:      folder.Icon,
		UpdatedAt: folder.UpdatedAt,
		Notes:     []app.WorkspaceTreeNote{},
		Children:  []app.WorkspaceTreeFolder{},
	}

	if notes, ok := notesByFolder[folder.ID]; ok {
		for _, note := range notes {
			result.Notes = append(result.Notes, app.WorkspaceTreeNote{
				Id:        note.ID,
				Name:      note.Name,
				Icon:      note.Icon,
				UpdatedAt: note.UpdatedAt,
			})
		}
	}

	for _, childFolder := range folderMap {
		if childFolder.ParentID != nil && *childFolder.ParentID == folder.ID {
			childTree := r.buildFolderTreeFromMap(*childFolder, folderMap, notesByFolder)
			result.Children = append(result.Children, *childTree)
		}
	}

	return &result
}

func (r *ReadModel) ShowTrash(ctx context.Context, q *app.ShowTrash) (*app.Trash, error) {
	trashedNotes, err := r.queries.GetTrashedNotesByWorkspaceID(ctx, q.WorkspaceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	trashedFolders, err := r.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &q.WorkspaceID,
		TrashedBy:   domain.TrashedByPurpose.String(),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	notes := make([]app.TrashedNote, len(trashedNotes))
	for i, note := range trashedNotes {
		notes[i] = app.TrashedNote{
			Id:        note.ID,
			Name:      note.Name,
			TrashedBy: domain.TrashedByPurpose,
			TrashedAt: *note.TrashedAt,
		}
	}

	folders := make([]app.TrashedFolder, len(trashedFolders))
	for i, folder := range trashedFolders {
		folders[i] = app.TrashedFolder{
			Id:        folder.ID,
			Name:      folder.Name,
			TrashedBy: domain.TrashedByPurpose,
			TrashedAt: *folder.TrashedAt,
		}
	}

	return &app.Trash{
		Notes:   notes,
		Folders: folders,
	}, nil
}

func (r *ReadModel) GetNoteGraph(ctx context.Context, q *app.GetNoteGraph) (*app.Graph, error) {
	workspaceID, err := r.queries.GetWorkspaceIDByNoteID(ctx, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNoteNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	graphRow, err := r.queries.GetNoteGraph(ctx, &pgsqlc.GetNoteGraphParams{
		WorkspaceID: workspaceID,
		StartNodeID: q.ID.String(),
		MaxDepth:    int32(q.Depth),
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	type tempNode struct {
		Id   string `json:"Id"`
		Name string `json:"Name"`
		Type string `json:"Type"`
	}
	type tempLink struct {
		Source string `json:"Source"`
		Target string `json:"Target"`
	}

	var nodes []tempNode
	var links []tempLink

	if graphRow.Nodes != nil {
		if nodesBytes, ok := graphRow.Nodes.([]byte); ok {
			if err := json.Unmarshal(nodesBytes, &nodes); err != nil {
				return nil, toDomainError(err)
			}
		}
	}

	if graphRow.Links != nil {
		if linksBytes, ok := graphRow.Links.([]byte); ok {
			if err := json.Unmarshal(linksBytes, &links); err != nil {
				return nil, toDomainError(err)
			}
		}
	}

	result := &app.Graph{
		Nodes: make([]app.GraphNode, len(nodes)),
		Links: make([]app.GraphLink, len(links)),
	}

	for i, node := range nodes {
		result.Nodes[i] = app.GraphNode{
			Id:   node.Id,
			Name: node.Name,
			Type: app.GraphNodeType(node.Type),
		}
	}

	for i, link := range links {
		result.Links[i] = app.GraphLink{
			Source: link.Source,
			Target: link.Target,
		}
	}

	return result, nil
}

func (r *ReadModel) GetNoteLinks(ctx context.Context, q *app.GetNoteLinks) (*app.NoteLinkResult, error) {
	_, err := r.queries.GetNote(ctx, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNoteNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	result := app.NoteLinkResult{
		OutgoingLinks: []app.NoteLink{},
		Backlinks:     []app.NoteLink{},
	}

	if q.OutgoingLinks {
		outgoingLinks, err := r.queries.GetNoteOutgoingLinks(ctx, q.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}

		if len(outgoingLinks) > 0 {
			outgoingNotes, err := r.queries.GetNotes(ctx, outgoingLinks)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, toDomainError(err)
			}
			for _, linkedNote := range outgoingNotes {
				result.OutgoingLinks = append(result.OutgoingLinks, app.NoteLink{
					Id:   linkedNote.ID,
					Name: linkedNote.Name,
					Icon: linkedNote.Icon,
				})
			}
		}
	}

	if q.Backlinks {
		backlinks, err := r.queries.GetNoteBacklinks(ctx, q.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}

		if len(backlinks) > 0 {
			backlinkNotes, err := r.queries.GetNotes(ctx, backlinks)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, toDomainError(err)
			}
			for _, linkedNote := range backlinkNotes {
				result.Backlinks = append(result.Backlinks, app.NoteLink{
					Id:   linkedNote.ID,
					Name: linkedNote.Name,
					Icon: linkedNote.Icon,
				})
			}
		}
	}

	return &result, nil
}

func (r *ReadModel) GetWorkspaceBySlug(ctx context.Context, q *app.GetWorkspaceBySlug) (*app.Workspace, error) {
	workspace, err := r.queries.GetWorkspace(ctx, &pgsqlc.GetWorkspaceParams{
		Slug: &q.Slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceBySlugNotFound(q.Slug, err)
		}
		return nil, toDomainError(err)
	}

	return &app.Workspace{
		Id:   workspace.ID,
		Slug: workspace.Slug,
		Name: workspace.Name,
	}, nil
}

func (r *ReadModel) GetWorkspaceGraph(ctx context.Context, q *app.GetWorkspaceGraph) (*app.Graph, error) {
	graphRow, err := r.queries.GetWorkspaceGraph(ctx, q.ID)
	if err != nil {
		return nil, toDomainError(err)
	}

	type tempNode struct {
		Id   string `json:"Id"`
		Name string `json:"Name"`
		Type string `json:"Type"`
	}
	type tempLink struct {
		Source string `json:"Source"`
		Target string `json:"Target"`
	}

	var nodes []tempNode
	var links []tempLink

	if graphRow.Nodes != nil {
		if nodesBytes, ok := graphRow.Nodes.([]byte); ok {
			if err := json.Unmarshal(nodesBytes, &nodes); err != nil {
				return nil, toDomainError(err)
			}
		}
	}

	if graphRow.Links != nil {
		if linksBytes, ok := graphRow.Links.([]byte); ok {
			if err := json.Unmarshal(linksBytes, &links); err != nil {
				return nil, toDomainError(err)
			}
		}
	}

	result := &app.Graph{
		Nodes: make([]app.GraphNode, len(nodes)),
		Links: make([]app.GraphLink, len(links)),
	}

	for i, node := range nodes {
		result.Nodes[i] = app.GraphNode{
			Id:   node.Id,
			Name: node.Name,
			Type: app.GraphNodeType(node.Type),
		}
	}

	for i, link := range links {
		result.Links[i] = app.GraphLink{
			Source: link.Source,
			Target: link.Target,
		}
	}

	return result, nil
}

func (r *ReadModel) CheckWorkspaceSlugExists(ctx context.Context, q *app.CheckWorkspaceSlugExists) (*app.CheckWorkspaceSlugExistsResult, error) {
	exists, err := r.queries.CheckSlugExists(ctx, q.Slug)
	if err != nil {
		return nil, toDomainError(err)
	}

	return &app.CheckWorkspaceSlugExistsResult{
		Exists: exists,
	}, nil
}
