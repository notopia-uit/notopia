package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type ReadModel struct {
	pool *pgxpool.Pool
}

func NewReadModel(pool *pgxpool.Pool) *ReadModel {
	return &ReadModel{pool: pool}
}

var ProvideReadModel = NewReadModel

var (
	_ app.GetWorkspaceTreeReadModel     = (*ReadModel)(nil)
	_ app.GetNotesReadModel             = (*ReadModel)(nil)
	_ app.ShowTrashReadModel            = (*ReadModel)(nil)
	_ app.GetNoteGraphReadModel         = (*ReadModel)(nil)
	_ app.GetNoteLinksReadModel         = (*ReadModel)(nil)
	_ app.GetWorkspaceBySlugReadModel   = (*ReadModel)(nil)
	_ app.GetWorkspaceGraphReadModel    = (*ReadModel)(nil)
	_ app.CheckWorkspaceExistsReadModel = (*ReadModel)(nil)
)

func (r *ReadModel) GetWorkspaceTree(ctx context.Context, q *app.GetWorkspaceTree) (*app.WorkspaceTreeFolder, error) {
	queries := pgsqlc.New(r.pool)

	workspace, err := queries.GetWorkspaceByID(ctx, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceByIDNotFound(q.ID.String(), err)
		}
		return nil, toDomainError(err)
	}

	rootFolder, err := queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
		WorkspaceID:  &workspace.ID,
		IsRootFolder: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceRootFolderNotFound(workspace.Slug, err)
		}
		return nil, toDomainError(err)
	}

	tree, err := r.buildFolderTree(ctx, queries, *rootFolder)
	return tree, err
}

func (r *ReadModel) buildFolderTree(ctx context.Context, queries *pgsqlc.Queries, folder pgsqlc.Folder) (*app.WorkspaceTreeFolder, error) {
	result := app.WorkspaceTreeFolder{
		Id:        folder.ID,
		Name:      folder.Name,
		Icon:      folder.Icon,
		UpdatedAt: folder.UpdatedAt,
		Notes:     []app.WorkspaceTreeNote{},
		Children:  []app.WorkspaceTreeFolder{},
	}

	notes, err := queries.GetNotesByFolderID(ctx, folder.ID)
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

	children, err := queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		ParentID: &folder.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, child := range children {
		childTree, err := r.buildFolderTree(ctx, queries, *child)
		if err != nil {
			return nil, err
		}
		result.Children = append(result.Children, *childTree)
	}

	return &result, nil
}

func (r *ReadModel) GetNotes(ctx context.Context, q *app.GetNotes) (*app.Paginated[app.Note], error) {
	queries := pgsqlc.New(r.pool)

	folder, err := queries.GetFolder(ctx, &pgsqlc.GetFolderParams{
		ID: &q.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrFolderNotFound(q.ID, err)
		}
		return nil, toDomainError(err)
	}

	notes, err := queries.GetNotesByFolderID(ctx, folder.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	noteIDs := make([]uuid.UUID, len(notes))
	for i, note := range notes {
		noteIDs[i] = note.ID
	}

	backlinksMap := make(map[uuid.UUID]int)
	outgoingLinksMap := make(map[uuid.UUID]int)

	if len(noteIDs) > 0 {
		for _, noteID := range noteIDs {
			backlinks, err := queries.GetNoteBacklinks(ctx, noteID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, toDomainError(err)
			}
			backlinksMap[noteID] = len(backlinks)
		}

		outgoingLinks, err := queries.GetNotesOutgoingLinks(ctx, noteIDs)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}
		for _, link := range outgoingLinks {
			outgoingLinksMap[link.SourceID]++
		}
	}

	appNotes := make([]app.Note, len(notes))
	for i, note := range notes {
		appNotes[i] = app.Note{
			Id:                 note.ID,
			Name:               note.Name,
			Icon:               note.Icon,
			Tags:               note.Tags,
			FolderId:           note.FolderID,
			BacklinksCount:     backlinksMap[note.ID],
			OutgoingLinksCount: outgoingLinksMap[note.ID],
			UpdatedAt:          note.UpdatedAt,
		}
	}

	pagination := app.Pagination{
		Page:       1,
		Limit:      len(appNotes),
		Total:      len(appNotes),
		TotalPages: 1,
	}

	return &app.Paginated[app.Note]{
		Data:       appNotes,
		Pagination: pagination,
	}, nil
}

func (r *ReadModel) ShowTrash(ctx context.Context, q *app.ShowTrash) (*app.Trash, error) {
	queries := pgsqlc.New(r.pool)

	workspaceID, err := queries.GetWorkspaceIDBySlug(ctx, q.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrWorkspaceBySlugNotFound(q.Slug, err)
		}
		return nil, toDomainError(err)
	}

	trashedNotes, err := queries.GetTrashedNotesByWorkspaceID(ctx, workspaceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	trashedFolders, err := queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &workspaceID,
		TrashedBy:   "purpose",
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	notes := make([]app.TrashedNote, len(trashedNotes))
	for i, note := range trashedNotes {
		trashedBy := domain.TrashedByPurpose
		if note.TrashedBy != nil {
			trashedBy = domain.TrashedBy(*note.TrashedBy)
		}
		notes[i] = app.TrashedNote{
			Id:        note.ID,
			Name:      note.Name,
			TrashedBy: trashedBy,
			TrashedAt: *note.TrashedAt,
		}
	}

	folders := make([]app.TrashedFolder, len(trashedFolders))
	for i, folder := range trashedFolders {
		trashedBy := domain.TrashedByPurpose
		if folder.TrashedBy != nil {
			trashedBy = domain.TrashedBy(*folder.TrashedBy)
		}
		folders[i] = app.TrashedFolder{
			Id:        folder.ID,
			Name:      folder.Name,
			TrashedBy: trashedBy,
			TrashedAt: *folder.TrashedAt,
		}
	}

	return &app.Trash{
		Notes:   notes,
		Folders: folders,
	}, nil
}

func (r *ReadModel) GetNoteGraph(ctx context.Context, q *app.GetNoteGraph) (*app.Graph, error) {
	queries := pgsqlc.New(r.pool)

	note, err := queries.GetNote(ctx, q.ID)
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

	outgoingLinks, err := queries.GetNoteOutgoingLinks(ctx, q.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, targetID := range outgoingLinks {
		targetIDStr := targetID.String()
		if !nodeMap[targetIDStr] {
			nodeMap[targetIDStr] = true
			targetNote, err := queries.GetNote(ctx, targetID)
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

	backlinks, err := queries.GetNoteBacklinks(ctx, q.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	for _, sourceID := range backlinks {
		sourceIDStr := sourceID.String()
		if !nodeMap[sourceIDStr] {
			nodeMap[sourceIDStr] = true
			sourceNote, err := queries.GetNote(ctx, sourceID)
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
	queries := pgsqlc.New(r.pool)

	_, err := queries.GetNote(ctx, q.ID)
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
		outgoingLinks, err := queries.GetNoteOutgoingLinks(ctx, q.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}

		if len(outgoingLinks) > 0 {
			outgoingNotes, err := queries.GetNotes(ctx, outgoingLinks)
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
		backlinks, err := queries.GetNoteBacklinks(ctx, q.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, toDomainError(err)
		}

		if len(backlinks) > 0 {
			backlinkNotes, err := queries.GetNotes(ctx, backlinks)
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
	queries := pgsqlc.New(r.pool)

	workspace, err := queries.GetWorkspaceBySlug(ctx, q.Slug)
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
	queries := pgsqlc.New(r.pool)

	folders, err := queries.GetFolders(ctx, &pgsqlc.GetFoldersParams{
		WorkspaceID: &q.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toDomainError(err)
	}

	folderIDs := make([]uuid.UUID, len(folders))
	for i, folder := range folders {
		folderIDs[i] = folder.ID
	}

	allNotes, err := queries.GetNotes(ctx, folderIDs)
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
		allLinks, err := queries.GetNotesOutgoingLinks(ctx, noteIDs)
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
	queries := pgsqlc.New(r.pool)

	exists, err := queries.CheckSlugExists(ctx, q.Slug)
	if err != nil {
		return nil, toDomainError(err)
	}

	return &app.CheckWorkspaceExistsResult{
		Exists: exists,
	}, nil
}
