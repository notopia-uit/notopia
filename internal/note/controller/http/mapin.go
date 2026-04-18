package http

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/authorization/errs"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

// This is partially, not have user name in
func toWorkspaceMemberUpdate(wmu *note.WorkspaceMember) (app.WorkspaceMemberUpdate, error) {
	role, err := toWorkspaceRole(wmu.Role)
	if err != nil {
		return app.WorkspaceMemberUpdate{}, err
	}
	return app.WorkspaceMemberUpdate{
		ID:   wmu.Id,
		Role: role,
	}, nil
}

func toWorkspaceMemberUpdates(wmus []note.WorkspaceMember) ([]app.WorkspaceMemberUpdate, error) {
	updates := make([]app.WorkspaceMemberUpdate, 0, len(wmus))
	for _, wmu := range wmus {
		update, err := toWorkspaceMemberUpdate(&wmu)
		if err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	return updates, nil
}

func toWorkspaceRole(r note.WorkspaceRole) (app.WorkspaceRole, error) {
	switch r {
	case note.Owner:
		return app.WorkspaceRoleOwner, nil
	case note.Editor:
		return app.WorkspaceRoleEditor, nil
	case note.Viewer:
		return app.WorkspaceRoleViewer, nil
	}
	return 0, errs.NewInvalid(fmt.Sprintf("invalid workspace role: %v", r))
}
