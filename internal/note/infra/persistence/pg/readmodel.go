package pg

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
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
	_ app.GetNoteReadModel                  = (*ReadModel)(nil)
)

func (r *ReadModel) GetWorkspaceTree(ctx context.Context, q *app.GetWorkspaceTree) (*app.WorkspaceTreeFolder, errs.Error) {
	workspace, err := r.queries.GetWorkspace(ctx, &pgsqlc.GetWorkspaceParams{
		ID: &q.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	rootFolder, err := r.queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
		WorkspaceID:  &workspace.ID,
		IsRootFolder: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceRootFolderNotFound(workspace.ID, err)
		}
		return nil, toDomainError(err)
	}

	allFolders, err := r.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &workspace.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	allNotes, err := r.queries.GetNotesInWorkspace(ctx, &pgsqlc.GetNotesInWorkspaceParams{
		WorkspaceID: workspace.ID,
	})
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
		ID:        folder.ID,
		Name:      folder.Name,
		Icon:      folder.Icon,
		UpdatedAt: folder.UpdatedAt,
		Notes:     []*app.WorkspaceTreeNote{},
		Children:  []*app.WorkspaceTreeFolder{},
	}

	if notes, ok := notesByFolder[folder.ID]; ok {
		for _, note := range notes {
			result.Notes = append(result.Notes, &app.WorkspaceTreeNote{
				ID:        note.ID,
				Name:      note.Name,
				Icon:      note.Icon,
				UpdatedAt: note.UpdatedAt,
			})
		}
	}

	for _, childFolder := range folderMap {
		if childFolder.ParentID != nil && *childFolder.ParentID == folder.ID {
			childTree := r.buildFolderTreeFromMap(*childFolder, folderMap, notesByFolder)
			result.Children = append(result.Children, childTree)
		}
	}

	return &result
}

func (r *ReadModel) ShowTrash(ctx context.Context, q *app.ShowTrash) (*app.Trash, errs.Error) {
	trashedNotes, err := r.queries.GetTrashedNotesByWorkspaceID(ctx, q.WorkspaceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	trashedFolders, err := r.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &q.WorkspaceID,
		TrashedBy:   new(domain.TrashedByPurpose.String()),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	notes := make([]*app.TrashedNote, len(trashedNotes))
	for i, note := range trashedNotes {
		notes[i] = &app.TrashedNote{
			ID:        note.ID,
			Name:      note.Name,
			TrashedBy: domain.TrashedByPurpose,
			TrashedAt: *note.TrashedAt,
		}
	}

	folders := make([]*app.TrashedFolder, len(trashedFolders))
	for i, folder := range trashedFolders {
		folders[i] = &app.TrashedFolder{
			ID:        folder.ID,
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

func (r *ReadModel) GetNoteLinks(ctx context.Context, q *app.GetNoteLinks) (*app.NoteLinkResult, errs.Error) {
	_, err := r.queries.GetNote(ctx, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewNoteNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	result := app.NoteLinkResult{
		OutgoingLinks: []*app.NoteLink{},
		Backlinks:     []*app.NoteLink{},
	}

	if q.OutgoingLinks {
		outgoingLinks, err := r.queries.GetNoteOutgoingLinks(ctx, &pgsqlc.GetNoteOutgoingLinksParams{
			SourceID: &q.ID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}

		if len(outgoingLinks) > 0 {
			outgoingNotes, err := r.queries.GetNotes(ctx, outgoingLinks)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, toDomainError(err)
			}
			for _, linkedNote := range outgoingNotes {
				result.OutgoingLinks = append(result.OutgoingLinks, &app.NoteLink{
					ID:   linkedNote.ID,
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
				result.Backlinks = append(result.Backlinks, &app.NoteLink{
					ID:   linkedNote.ID,
					Name: linkedNote.Name,
					Icon: linkedNote.Icon,
				})
			}
		}
	}

	return &result, nil
}

func (r *ReadModel) GetWorkspaceBySlug(ctx context.Context, q *app.GetWorkspaceBySlug) (*app.Workspace, errs.Error) {
	workspace, err := r.queries.GetWorkspace(ctx, &pgsqlc.GetWorkspaceParams{
		Slug: &q.Slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewWorkspaceBySlugNotFound(q.Slug, err)
		}
		return nil, toDomainError(err)
	}

	return &app.Workspace{
		ID:   workspace.ID,
		Slug: workspace.Slug,
		Name: workspace.Name,
	}, nil
}

func (r *ReadModel) CheckWorkspaceSlugExists(ctx context.Context, q *app.CheckWorkspaceSlugExists) (*app.CheckWorkspaceSlugExistsResult, errs.Error) {
	exists, err := r.queries.CheckSlugExists(ctx, q.Slug)
	if err != nil {
		return nil, toDomainError(err)
	}

	return &app.CheckWorkspaceSlugExistsResult{
		Exists: exists,
	}, nil
}

func (r *ReadModel) GetWorkspaceGraph(ctx context.Context, q *app.GetWorkspaceGraph) (*app.Graph, errs.Error) {
	notes, err := r.queries.GetNotesInWorkspace(ctx, &pgsqlc.GetNotesInWorkspaceParams{
		WorkspaceID: q.ID,
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	links, err := r.queries.GetNoteLinksInWorkspace(ctx, q.ID)
	if err != nil {
		return nil, toDomainError(err)
	}

	reachableIDs := make(map[string]bool, len(notes))
	for _, n := range notes {
		reachableIDs[n.ID.String()] = true
	}

	return buildGraph(notes, links, reachableIDs), nil
}

func (r *ReadModel) GetNoteGraph(ctx context.Context, q *app.GetNoteGraph) (*app.Graph, errs.Error) {
	workspaceID, err := r.queries.GetWorkspaceIDByNoteID(ctx, q.ID)
	if err != nil {
		return nil, toDomainError(err)
	}

	notes, err := r.queries.GetNotesInWorkspace(ctx, &pgsqlc.GetNotesInWorkspaceParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, toDomainError(err)
	}

	links, err := r.queries.GetNoteLinksInWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, toDomainError(err)
	}

	// 2. Build Adjacency List for Traversal (Bidirectional for backlinks and tags)
	adj := make(map[string][]string)
	for _, n := range notes {
		for _, tag := range n.Tags {
			tagID := "#" + tag
			adj[n.ID.String()] = append(adj[n.ID.String()], tagID) // Note -> Tag
			adj[tagID] = append(adj[tagID], n.ID.String())         // Tag -> Note
		}
	}
	for _, l := range links {
		adj[l.SourceID.String()] = append(adj[l.SourceID.String()], l.TargetID.String()) // Note -> Note
		adj[l.TargetID.String()] = append(adj[l.TargetID.String()], l.SourceID.String()) // Backlink: Note <- Note
	}

	// 3. BFS Traversal to find reachable nodes within maxDepth
	reachableIDs := make(map[string]bool)
	type queueItem struct {
		id    string
		depth int
	}

	startID := q.ID.String()
	queue := []queueItem{{id: startID, depth: 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if reachableIDs[curr.id] {
			continue
		}
		reachableIDs[curr.id] = true

		if curr.depth < q.Depth {
			for _, neighbor := range adj[curr.id] {
				if !reachableIDs[neighbor] {
					queue = append(queue, queueItem{id: neighbor, depth: curr.depth + 1})
				}
			}
		}
	}

	return buildGraph(notes, links, reachableIDs), nil
}

func calculateGraphWeight(size, minSize, maxSize int32) *float64 {
	var w float64
	if maxSize == minSize {
		w = 1
	} else {
		w = float64(size-minSize) / float64(maxSize-minSize)
	}
	return &w
}

func buildGraph(notes []*pgsqlc.Note, links []*pgsqlc.NoteLink, reachableIDs map[string]bool) *app.Graph {
	var minSize int32 = math.MaxInt32
	var maxSize int32 = -1
	reachableNotesMap := make(map[uuid.UUID]*pgsqlc.Note)

	for _, n := range notes {
		if reachableIDs[n.ID.String()] {
			reachableNotesMap[n.ID] = n
			if n.Size < minSize {
				minSize = n.Size
			}
			if n.Size > maxSize {
				maxSize = n.Size
			}
		}
	}

	var graphNodes []*app.GraphNode
	var graphLinks []*app.GraphLink
	tagsAdded := make(map[string]bool)

	// 2. Build Nodes (Notes and Tags)
	for _, n := range reachableNotesMap {
		// Add Note Node
		graphNodes = append(graphNodes, &app.GraphNode{
			ID:     n.ID.String(),
			Name:   n.Name,
			Type:   app.GraphNodeTypeNote,
			Weight: calculateGraphWeight(n.Size, minSize, maxSize),
		})

		for _, tag := range n.Tags {
			tagID := "#" + tag

			if reachableIDs[tagID] {
				if !tagsAdded[tagID] {
					graphNodes = append(graphNodes, &app.GraphNode{
						ID:     tagID,
						Name:   tag,
						Type:   app.GraphNodeTypeTag,
						Weight: nil,
					})
					tagsAdded[tagID] = true
				}

				// Add structural link for Note -> Tag
				graphLinks = append(graphLinks, &app.GraphLink{
					Source: n.ID.String(),
					Target: tagID,
				})
			}
		}
	}

	// 3. Build Note -> Note Links (filtering out unreachable ones)
	for _, l := range links {
		if reachableIDs[l.SourceID.String()] && reachableIDs[l.TargetID.String()] {
			graphLinks = append(graphLinks, &app.GraphLink{
				Source: l.SourceID.String(),
				Target: l.TargetID.String(),
			})
		}
	}

	return &app.Graph{
		Nodes: graphNodes,
		Links: graphLinks,
	}
}

func (r *ReadModel) GetNote(ctx context.Context, q *app.GetNote) (*app.Note, errs.Error) {
	return nil, nil
}
