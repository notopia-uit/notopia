package app_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserWorkspaceItemPermissionsHandler(t *testing.T) {
	e, err := GetLocalEnforcer(true)
	require.NoError(t, err, "Failed to create enforcer")

	handler := app.NewGetUserWorkspaceItemPermissionsHandler(e)

	tests := []struct {
		name        string
		userID      string
		workspaceID string
		expected    app.WorkspaceItemPermissions
	}{
		{"W111-Owner: All permissions", "111", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}},
		{"W111-Editor: Read, Write, Delete", "112", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}},
		{"W111-Viewer: Read only", "110", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissions{Read: true, Write: false, Delete: false}},
		{"W112-Owner: All permissions", "112", "00000000-0000-0000-0000-000000000112", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}},
		{"W112-Editor: Read, Write, Delete", "111", "00000000-0000-0000-0000-000000000112", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}},
		{"W112-Stranger: No permissions", "110", "00000000-0000-0000-0000-000000000112", app.WorkspaceItemPermissions{Read: false, Write: false, Delete: false}},
		{"W110-Owner: All permissions", "110", "00000000-0000-0000-0000-000000000110", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}},
		{"W110-Stranger user 111: No permissions", "111", "00000000-0000-0000-0000-000000000110", app.WorkspaceItemPermissions{Read: false, Write: false, Delete: false}},
		{"W110-Stranger user 112: No permissions", "112", "00000000-0000-0000-0000-000000000110", app.WorkspaceItemPermissions{Read: false, Write: false, Delete: false}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			workspaceID := uuid.MustParse(tc.workspaceID)
			perms, err := handler.Handle(app.GetUserWorkspaceItemPermissions{
				UserID:      tc.userID,
				WorkspaceID: workspaceID,
			})
			require.NoError(t, err, "Handler threw an error")
			assert.Equal(t, tc.expected, *perms)
		})
	}
}
