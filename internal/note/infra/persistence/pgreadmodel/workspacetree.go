package pgreadmodel

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type WorkspaceTree struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspaceTreeReadModel = (*WorkspaceTree)(nil)

func NewWorkspaceTree(queries *pgsqlc.Queries) *WorkspaceTree {
	return &WorkspaceTree{queries: queries}
}

var ProvideWorkspaceTree = NewWorkspaceTree

func (h *WorkspaceTree) Handle(ctx context.Context, p *app.GetWorkspaceTreeReadModelParams) (app.WorkspaceTreeFolder, error) {
	var rootFolderID uuid.UUID

	if p.RootFolderID != uuid.Nil {
		rootFolderID = p.RootFolderID
	} else {
		rootFolderIDs, err := h.queries.ReadGetRootFolderIDsByWorkspaceID(ctx, p.WorkspaceID)
		if err != nil {
			return app.WorkspaceTreeFolder{}, toErr(err)
		}
		if len(rootFolderIDs) == 0 {
			return app.WorkspaceTreeFolder{}, errs.NewWorkspaceRootFolderNotFound(p.WorkspaceID, pgx.ErrNoRows)
		}
		rootFolderID = rootFolderIDs[0]
	}

	rootFolder, err := h.queries.ReadGetFolderByID(ctx, rootFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.WorkspaceTreeFolder{}, errs.NewFolderNotFound(rootFolderID, err)
		}
		return app.WorkspaceTreeFolder{}, toErr(err)
	}

	var depth *int32
	if p.Depth != 0 {
		depth = new(int32(p.Depth))
	}
	recursiveFolders, err := h.queries.ReadGetRecursiveFolderByParentID(ctx, &pgsqlc.ReadGetRecursiveFolderByParentIDParams{
		ParentID:       rootFolderID,
		Depth:          depth,
		IncludeTrashed: p.IncludeTrashed,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return app.WorkspaceTreeFolder{}, toErr(err)
	}

	var folderIDs []uuid.UUID
	folderIDs = append(folderIDs, rootFolderID)
	folderMap := make(map[uuid.UUID]*pgsqlc.ReadGetRecursiveFolderByParentIDRow)
	childrenByParentID := make(map[uuid.UUID][]*pgsqlc.ReadGetRecursiveFolderByParentIDRow)

	for _, folder := range recursiveFolders {
		folderIDs = append(folderIDs, folder.ID)
		folderMap[folder.ID] = folder
	}

	// Precompute parent -> children mapping for O(1) access
	for _, folder := range recursiveFolders {
		if folder.ParentID != nil {
			childrenByParentID[*folder.ParentID] = append(childrenByParentID[*folder.ParentID], folder)
		}
	}

	for _, children := range childrenByParentID {
		sortFolders(children, p.Sort)
	}

	allNotes, err := h.queries.ReadGetNotesByFolderIDs(ctx, pgsqlc.ReadGetNotesByFolderIDsParams{
		FolderIds:    folderIDs,
		ExcludeTrash: !p.IncludeTrashed,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return app.WorkspaceTreeFolder{}, toErr(err)
	}

	notesByFolder := make(map[uuid.UUID][]*pgsqlc.Note)
	for _, note := range allNotes {
		notesByFolder[note.FolderID] = append(notesByFolder[note.FolderID], note)
	}

	for _, notes := range notesByFolder {
		sortNotes(notes, p.Sort)
	}

	tree := h.buildFolderTree(
		rootFolder.ID,
		rootFolder.Name,
		rootFolder.Icon,
		rootFolder.UpdatedAt,
		childrenByParentID,
		notesByFolder,
	)
	return tree, nil
}

func (h *WorkspaceTree) buildFolderTree(
	folderID uuid.UUID,
	folderName string,
	folderIcon *string,
	updatedAt time.Time,
	childrenByParentID map[uuid.UUID][]*pgsqlc.ReadGetRecursiveFolderByParentIDRow,
	notesByFolder map[uuid.UUID][]*pgsqlc.Note,
) app.WorkspaceTreeFolder {
	var icon string
	if folderIcon != nil {
		icon = *folderIcon
	}
	result := app.WorkspaceTreeFolder{
		ID:        folderID,
		Name:      folderName,
		Icon:      icon,
		UpdatedAt: updatedAt,
		Notes:     []app.WorkspaceTreeNote{},
		Children:  []app.WorkspaceTreeFolder{},
	}

	if notes, ok := notesByFolder[folderID]; ok {
		result.Notes = make([]app.WorkspaceTreeNote, len(notes))
		for i, note := range notes {
			var noteIcon string
			if note.Icon != nil {
				noteIcon = *note.Icon
			}
			result.Notes[i] = app.WorkspaceTreeNote{
				ID:        note.ID,
				Name:      note.Name,
				Icon:      noteIcon,
				UpdatedAt: note.UpdatedAt,
			}
		}
	}

	// Only iterate through direct children, not the entire folderMap
	if children, ok := childrenByParentID[folderID]; ok {
		result.Children = make([]app.WorkspaceTreeFolder, len(children))
		for i, childFolder := range children {
			result.Children[i] = h.buildFolderTree(
				childFolder.ID,
				childFolder.Name,
				childFolder.Icon,
				childFolder.UpdatedAt,
				childrenByParentID,
				notesByFolder,
			)
		}
	}

	return result
}

func sortFolders(folders []*pgsqlc.ReadGetRecursiveFolderByParentIDRow, sortBy app.GetWorkspaceTreeSort) {
	if sortBy.Name != app.SortOrderUnspecified || sortBy.CreatedAt != app.SortOrderUnspecified || sortBy.UpdatedAt != app.SortOrderUnspecified {
		sort.SliceStable(folders, func(i, j int) bool {
			if sortBy.Name != app.SortOrderUnspecified {
				if folders[i].Name != folders[j].Name {
					if sortBy.Name == app.SortOrderAsc {
						return folders[i].Name < folders[j].Name
					}
					return folders[i].Name > folders[j].Name
				}
			}
			if sortBy.CreatedAt != app.SortOrderUnspecified {
				if !folders[i].CreatedAt.Equal(folders[j].CreatedAt) {
					if sortBy.CreatedAt == app.SortOrderAsc {
						return folders[i].CreatedAt.Before(folders[j].CreatedAt)
					}
					return folders[i].CreatedAt.After(folders[j].CreatedAt)
				}
			}
			if sortBy.UpdatedAt != app.SortOrderUnspecified {
				if sortBy.UpdatedAt == app.SortOrderAsc {
					return folders[i].UpdatedAt.Before(folders[j].UpdatedAt)
				}
				return folders[i].UpdatedAt.After(folders[j].UpdatedAt)
			}
			return false
		})
	} else {
		sort.Slice(folders, func(i, j int) bool {
			return folders[i].Name < folders[j].Name
		})
	}
}

func sortNotes(notes []*pgsqlc.Note, sortBy app.GetWorkspaceTreeSort) {
	if sortBy.Name != app.SortOrderUnspecified || sortBy.CreatedAt != app.SortOrderUnspecified || sortBy.UpdatedAt != app.SortOrderUnspecified {
		sort.SliceStable(notes, func(i, j int) bool {
			if sortBy.Name != app.SortOrderUnspecified {
				if notes[i].Name != notes[j].Name {
					if sortBy.Name == app.SortOrderAsc {
						return notes[i].Name < notes[j].Name
					}
					return notes[i].Name > notes[j].Name
				}
			}
			if sortBy.CreatedAt != app.SortOrderUnspecified {
				if !notes[i].CreatedAt.Equal(notes[j].CreatedAt) {
					if sortBy.CreatedAt == app.SortOrderAsc {
						return notes[i].CreatedAt.Before(notes[j].CreatedAt)
					}
					return notes[i].CreatedAt.After(notes[j].CreatedAt)
				}
			}
			if sortBy.UpdatedAt != app.SortOrderUnspecified {
				if sortBy.UpdatedAt == app.SortOrderAsc {
					return notes[i].UpdatedAt.Before(notes[j].UpdatedAt)
				}
				return notes[i].UpdatedAt.After(notes[j].UpdatedAt)
			}
			return false
		})
	} else {
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Name < notes[j].Name
		})
	}
}
