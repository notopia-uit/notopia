package pgreadmodel

import (
	"context"
	"errors"
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

func (h *WorkspaceTree) GetWorkspaceTree(ctx context.Context, q *app.GetWorkspaceTree) (*app.WorkspaceTreeFolder, error) {
	var rootFolderID uuid.UUID

	if q.RootFolderID != uuid.Nil {
		rootFolderID = q.RootFolderID
	} else {
		rootFolderIDs, err := h.queries.ReadGetRootFolderIDsByWorkspaceID(ctx, q.WorkspaceID)
		if err != nil {
			return nil, toErr(err)
		}
		if len(rootFolderIDs) == 0 {
			return nil, errs.NewWorkspaceRootFolderNotFound(q.WorkspaceID, pgx.ErrNoRows)
		}
		rootFolderID = rootFolderIDs[0]
	}

	rootFolder, err := h.queries.ReadGetFolderByID(ctx, rootFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewFolderNotFound(rootFolderID, err)
		}
		return nil, toErr(err)
	}

	var depth *int32
	if q.Depth != 0 {
		depth = new(int32(q.Depth))
	}
	recursiveFolders, err := h.queries.ReadGetRecursiveFolderByParentID(ctx, &pgsqlc.ReadGetRecursiveFolderByParentIDParams{
		ParentID:       rootFolderID,
		Depth:          depth,
		IncludeTrashed: q.IncludeTrashed,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	var folderIDs []uuid.UUID
	folderIDs = append(folderIDs, rootFolderID)
	folderMap := make(map[uuid.UUID]*pgsqlc.ReadGetRecursiveFolderByParentIDRow)
	for _, folder := range recursiveFolders {
		folderIDs = append(folderIDs, folder.ID)
		folderMap[folder.ID] = folder
	}

	allNotes, err := h.queries.ReadGetNotesByFolderIDs(ctx, pgsqlc.ReadGetNotesByFolderIDsParams{
		FolderIds:    folderIDs,
		ExcludeTrash: !q.IncludeTrashed,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	notesByFolder := make(map[uuid.UUID][]*pgsqlc.Note)
	for _, note := range allNotes {
		notesByFolder[note.FolderID] = append(notesByFolder[note.FolderID], note)
	}

	tree := h.buildFolderTree(
		rootFolder.ID,
		rootFolder.Name,
		rootFolder.Icon,
		rootFolder.UpdatedAt,
		folderMap,
		notesByFolder,
	)
	return tree, nil
}

func (h *WorkspaceTree) buildFolderTree(
	folderID uuid.UUID,
	folderName string,
	folderIcon *string,
	updatedAt time.Time,
	folderMap map[uuid.UUID]*pgsqlc.ReadGetRecursiveFolderByParentIDRow,
	notesByFolder map[uuid.UUID][]*pgsqlc.Note,
) *app.WorkspaceTreeFolder {
	var icon string
	if folderIcon != nil {
		icon = *folderIcon
	}
	result := app.WorkspaceTreeFolder{
		ID:        folderID,
		Name:      folderName,
		Icon:      icon,
		UpdatedAt: updatedAt,
		Notes:     []*app.WorkspaceTreeNote{},
		Children:  []*app.WorkspaceTreeFolder{},
	}

	if notes, ok := notesByFolder[folderID]; ok {
		for _, note := range notes {
			var noteIcon string
			if note.Icon != nil {
				noteIcon = *note.Icon
			}
			result.Notes = append(result.Notes, &app.WorkspaceTreeNote{
				ID:        note.ID,
				Name:      note.Name,
				Icon:      noteIcon,
				UpdatedAt: note.UpdatedAt,
			})
		}
	}

	for _, childFolder := range folderMap {
		if childFolder.ParentID != nil && *childFolder.ParentID == folderID {
			childTree := h.buildFolderTree(
				childFolder.ID,
				childFolder.Name,
				childFolder.Icon,
				childFolder.UpdatedAt,
				folderMap,
				notesByFolder,
			)
			result.Children = append(result.Children, childTree)
		}
	}

	return &result
}
