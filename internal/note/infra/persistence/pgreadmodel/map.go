package pgreadmodel

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

func toAppFolder(f *pgsqlc.Folder) (app.Folder, error) {
	trashed, err := toAppTrashed(f.TrashedAt, f.TrashedBy)
	if err != nil {
		return app.Folder{}, err
	}
	var icon string
	if f.Icon != nil {
		icon = *f.Icon
	}
	var parentID uuid.UUID
	if f.ParentID != nil {
		parentID = *f.ParentID
	}
	return app.Folder{
		ID:          f.ID,
		Name:        f.Name,
		Trashed:     trashed,
		WorkspaceID: f.WorkspaceID,
		UpdatedAt:   f.UpdatedAt,
		Icon:        icon,
		ParentID:    parentID,
	}, nil
}

func toAppWorkspaces(workspaces []*pgsqlc.Workspace) []app.Workspace {
	appWorkspaces := make([]app.Workspace, len(workspaces))
	for i, w := range workspaces {
		appWorkspaces[i] = toAppWorkspace(w)
	}
	return appWorkspaces
}

func toAppWorkspace(w *pgsqlc.Workspace) app.Workspace {
	return app.Workspace{
		ID:   w.ID,
		Slug: w.Slug,
		Name: w.Name,
	}
}

func toAppTrashed(at *time.Time, by *string) (app.Trashed, error) {
	if at == nil && by == nil {
		return app.Trashed{}, nil
	}
	trashedBy, err := toAppTrashedBy(by)
	if err != nil {
		return app.Trashed{}, err
	}
	return app.Trashed{
		At: *at,
		By: trashedBy,
	}, nil
}

func toAppTrashedBy(trashedBy *string) (app.TrashedBy, error) {
	if trashedBy == nil {
		return app.TrashedByUnspecified, errs.NewPersistenceInternal("trashedBy is nil", nil)
	}
	switch *trashedBy {
	case string(pgsqlc.TrashedByPurpose):
		return app.TrashedByPurpose, nil
	case string(pgsqlc.TrashedByParent):
		return app.TrashedByParent, nil
	default:
		return app.TrashedByUnspecified, errs.NewPersistenceInternal(fmt.Sprintf("unknown trashedBy value: %s", *trashedBy), nil)
	}
}
