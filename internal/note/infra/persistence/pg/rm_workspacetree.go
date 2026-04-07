package pg

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

type GetWorkspaceTreeReadModel struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspaceTreeReadModel = (*GetWorkspaceTreeReadModel)(nil)

func NewGetWorkspaceTreeReadModel(queries *pgsqlc.Queries) *GetWorkspaceTreeReadModel {
	return &GetWorkspaceTreeReadModel{queries: queries}
}

var ProvideGetWorkspaceTreeReadModel = NewGetWorkspaceTreeReadModel

func (h *GetWorkspaceTreeReadModel) GetWorkspaceTree(ctx context.Context, q *app.GetWorkspaceTree) (*app.WorkspaceTreeFolder, error) {
	var rootFolderID uuid.UUID

	if q.RootFolderID != nil {
		rootFolderID = *q.RootFolderID
	} else {
		rootFolderIDs, err := h.queries.GetRootFolderIDsByWorkspaceID(ctx, q.WorkspaceID)
		if err != nil {
			return nil, toErr(err)
		}
		if len(rootFolderIDs) == 0 {
			return nil, errs.NewWorkspaceRootFolderNotFound(q.WorkspaceID, pgx.ErrNoRows)
		}
		rootFolderID = rootFolderIDs[0]
	}

	rootFolder, err := h.queries.GetFolder(ctx,
		//exhaustruct:ignore
		&pgsqlc.GetFolderParams{
			ID: rootFolderID,
		})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewFolderNotFound(rootFolderID, err)
		}
		return nil, toErr(err)
	}

	var depth *int32
	if q.Depth != nil {
		depth = new(int32(*q.Depth))
	}
	recursiveFolders, err := h.queries.GetRecursiveFolderByParentID(ctx, &pgsqlc.GetRecursiveFolderByParentIDParams{
		ParentID:       rootFolderID,
		Depth:          depth,
		IncludeTrashed: q.IncludeTrashed,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, toErr(err)
	}

	var folderIDs []uuid.UUID
	folderIDs = append(folderIDs, rootFolderID)
	folderMap := make(map[uuid.UUID]*pgsqlc.GetRecursiveFolderByParentIDRow)
	for _, folder := range recursiveFolders {
		folderIDs = append(folderIDs, folder.ID)
		folderMap[folder.ID] = folder
	}

	allNotes, err := h.queries.GetNotesByFolderIDs(ctx,
		//exhaustruct:ignore
		pgsqlc.GetNotesByFolderIDsParams{
			FolderIds:      folderIDs,
			IncludeTrashed: q.IncludeTrashed,
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

func (h *GetWorkspaceTreeReadModel) buildFolderTree(
	folderID uuid.UUID,
	folderName string,
	folderIcon *string,
	updatedAt time.Time,
	folderMap map[uuid.UUID]*pgsqlc.GetRecursiveFolderByParentIDRow,
	notesByFolder map[uuid.UUID][]*pgsqlc.Note,
) *app.WorkspaceTreeFolder {
	result := app.WorkspaceTreeFolder{
		ID:        folderID,
		Name:      folderName,
		Icon:      folderIcon,
		UpdatedAt: updatedAt,
		Notes:     []*app.WorkspaceTreeNote{},
		Children:  []*app.WorkspaceTreeFolder{},
	}

	if notes, ok := notesByFolder[folderID]; ok {
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
