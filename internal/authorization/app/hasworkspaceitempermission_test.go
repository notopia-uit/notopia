package app_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasWorkspaceItemPermissionHandler(t *testing.T) {
	e, err := GetLocalEnforcer(true)
	require.NoError(t, err, "Failed to create enforcer")

	handler := app.NewHasWorkspaceItemPermissionHandler(e)

	tests := []struct {
		name        string
		userID      string
		workspaceID string
		permission  app.WorkspaceItemPermission
		expected    bool
	}{
		{"W111-Owner: Read Note", "111", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionRead, true},
		{"W111-Owner: Write Note", "111", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionWrite, true},
		{"W111-Owner: Delete Folder", "111", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionDelete, true},
		{"W111-Editor: Read Note", "112", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionRead, true},
		{"W111-Editor: Write Note", "112", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionWrite, true},
		{"W111-Editor: Delete Note", "112", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionDelete, true},
		{"W111-Editor: Write Folder", "112", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionWrite, true},
		{"W111-Viewer: Read Note", "110", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionRead, true},
		{"W111-Viewer: CANNOT Write Note", "110", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionWrite, false},
		{"W111-Viewer: CANNOT Delete Folder", "110", "00000000-0000-0000-0000-000000000111", app.WorkspaceItemPermissionDelete, false},
		{"W112-Owner: Write Note", "112", "00000000-0000-0000-0000-000000000112", app.WorkspaceItemPermissionWrite, true},
		{"W112-Editor: Write Note", "111", "00000000-0000-0000-0000-000000000112", app.WorkspaceItemPermissionWrite, true},
		{"W112-Stranger: CANNOT Read Note", "110", "00000000-0000-0000-0000-000000000112", app.WorkspaceItemPermissionRead, false},
		{"W110-Stranger: User 111 CANNOT Read W110", "111", "00000000-0000-0000-0000-000000000110", app.WorkspaceItemPermissionRead, false},
		{"W110-Stranger: User 112 CANNOT Read W110", "112", "00000000-0000-0000-0000-000000000110", app.WorkspaceItemPermissionRead, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			workspaceID := uuid.MustParse(tc.workspaceID)
			ok, err := handler.Handle(app.HasWorkspaceItemPermission{
				UserID:      tc.userID,
				WorkspaceID: workspaceID,
				Permission:  tc.permission,
			})
			require.NoError(t, err, "Handler threw an error")
			assert.Equal(t, tc.expected, ok)
		})
	}
}
