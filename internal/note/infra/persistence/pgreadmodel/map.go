package pgreadmodel

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

func toAppWorkspaces(workspaces []*pgsqlc.Workspace) []*app.Workspace {
	appWorkspaces := make([]*app.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		appWorkspaces = append(appWorkspaces, toAppWorkspace(w))
	}
	return appWorkspaces
}

func toAppWorkspace(w *pgsqlc.Workspace) *app.Workspace {
	return &app.Workspace{
		ID:   w.ID,
		Slug: w.Slug,
		Name: w.Name,
	}
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
