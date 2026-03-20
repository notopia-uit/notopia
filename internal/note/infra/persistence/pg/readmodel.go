package pg

import (
	"context"
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
	_ app.GetWorkspaceTreeReadModel     = (*ReadModel)(nil)
	_ app.ShowTrashReadModel            = (*ReadModel)(nil)
	_ app.GetNoteGraphReadModel         = (*ReadModel)(nil)
	_ app.GetNoteLinksReadModel         = (*ReadModel)(nil)
	_ app.GetWorkspaceBySlugReadModel   = (*ReadModel)(nil)
	_ app.GetWorkspaceGraphReadModel    = (*ReadModel)(nil)
	_ app.CheckWorkspaceExistsReadModel = (*ReadModel)(nil)
)

func (r *ReadModel) GetWorkspaceTree(ctx context.Context, q *app.GetWorkspaceTree) (*app.WorkspaceTreeFolder, error) {
	workspace, err := r.queries.GetWorkspaceByID(ctx, q.ID)
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

	tree, err := r.buildFolderTree(ctx, *rootFolder)
	return tree, err
}

func (r *ReadModel) buildFolderTree(ctx context.Context, folder pgsqlc.Folder) (*app.WorkspaceTreeFolder, error) {
	result := app.WorkspaceTreeFolder{
		Id:        folder.ID,
		Name:      folder.Name,
		Icon:      folder.Icon,
		UpdatedAt: folder.UpdatedAt,
		Notes:     []app.WorkspaceTreeNote{},
		Children:  []app.WorkspaceTreeFolder{},
	}

	notes, err := r.queries.GetNotesByFolderID(ctx, folder.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, note := range notes {
		result.Notes = append(result.Notes, app.WorkspaceTreeNote{
			Id:        note.ID,
			Name:      note.Name,
			Icon:      note.Icon,
			UpdatedAt: note.UpdatedAt,
		})
	}

	children, err := r.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		ParentID: &folder.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, child := range children {
		childTree, err := r.buildFolderTree(ctx, *child)
		if err != nil {
			return nil, err
		}
		result.Children = append(result.Children, *childTree)
	}

	return &result, nil
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
	note, err := r.queries.GetNote(ctx, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNoteNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	nodeMap := make(map[string]bool)
	linkMap := make(map[string]bool)
	nodes := []app.GraphNode{}
	links := []app.GraphLink{}

	nodeID := q.ID.String()
	if !nodeMap[nodeID] {
		nodeMap[nodeID] = true
		nodes = append(nodes, app.GraphNode{
			Id:   nodeID,
			Name: note.Name,
			Type: "note",
		})
	}

	outgoingLinks, err := r.queries.GetNoteOutgoingLinks(ctx, q.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, targetID := range outgoingLinks {
		targetIDStr := targetID.String()
		if !nodeMap[targetIDStr] {
			nodeMap[targetIDStr] = true
			targetNote, err := r.queries.GetNote(ctx, targetID)
			if err == nil {
				nodes = append(nodes, app.GraphNode{
					Id:   targetIDStr,
					Name: targetNote.Name,
					Type: "note",
				})
			}
		}
		linkKey := nodeID + "->" + targetIDStr
		if !linkMap[linkKey] {
			linkMap[linkKey] = true
			links = append(links, app.GraphLink{
				Source: nodeID,
				Target: targetIDStr,
			})
		}
	}

	backlinks, err := r.queries.GetNoteBacklinks(ctx, q.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, sourceID := range backlinks {
		sourceIDStr := sourceID.String()
		if !nodeMap[sourceIDStr] {
			nodeMap[sourceIDStr] = true
			sourceNote, err := r.queries.GetNote(ctx, sourceID)
			if err == nil {
				nodes = append(nodes, app.GraphNode{
					Id:   sourceIDStr,
					Name: sourceNote.Name,
					Type: "note",
				})
			}
		}
		linkKey := sourceIDStr + "->" + nodeID
		if !linkMap[linkKey] {
			linkMap[linkKey] = true
			links = append(links, app.GraphLink{
				Source: sourceIDStr,
				Target: nodeID,
			})
		}
	}

	return &app.Graph{
		Nodes: nodes,
		Links: links,
	}, nil
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

	shouldFetchOutgoing := q.OutgoingLinks == nil || *q.OutgoingLinks
	shouldFetchBacklinks := q.Backlinks == nil || *q.Backlinks

	if shouldFetchOutgoing {
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

	if shouldFetchBacklinks {
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
	workspace, err := r.queries.GetWorkspaceBySlug(ctx, q.Slug)
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
	folders, err := r.queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &q.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	folderIDs := make([]uuid.UUID, len(folders))
	for i, folder := range folders {
		folderIDs[i] = folder.ID
	}

	allNotes, err := r.queries.GetNotes(ctx, folderIDs)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	noteIDs := make([]uuid.UUID, len(allNotes))
	for i, note := range allNotes {
		noteIDs[i] = note.ID
	}

	nodeMap := make(map[string]bool)
	linkMap := make(map[string]bool)
	nodes := []app.GraphNode{}
	links := []app.GraphLink{}

	for _, note := range allNotes {
		nodeID := note.ID.String()
		if !nodeMap[nodeID] {
			nodeMap[nodeID] = true
			nodes = append(nodes, app.GraphNode{
				Id:   nodeID,
				Name: note.Name,
				Type: "note",
			})
		}
	}

	if len(noteIDs) > 0 {
		allLinks, err := r.queries.GetNotesOutgoingLinks(ctx, noteIDs)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}

		for _, link := range allLinks {
			sourceID := link.SourceID.String()
			targetID := link.TargetID.String()

			if nodeMap[sourceID] && nodeMap[targetID] {
				linkKey := sourceID + "->" + targetID
				if !linkMap[linkKey] {
					linkMap[linkKey] = true
					links = append(links, app.GraphLink{
						Source: sourceID,
						Target: targetID,
					})
				}
			}
		}
	}

	return &app.Graph{
		Nodes: nodes,
		Links: links,
	}, nil
}

func (r *ReadModel) CheckWorkspaceExists(ctx context.Context, q *app.CheckWorkspaceExists) (*app.CheckWorkspaceExistsResult, error) {
	exists, err := r.queries.CheckSlugExists(ctx, q.Slug)
	if err != nil {
		return nil, toDomainError(err)
	}

	return &app.CheckWorkspaceExistsResult{
		Exists: exists,
	}, nil
}
