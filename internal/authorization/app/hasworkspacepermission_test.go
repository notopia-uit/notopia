package app_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasWorkspacePermissionHandler(t *testing.T) {
	e, err := GetLocalEnforcer(t, &GetLocalEnforcerParams{LoadTestPolicies: true})
	require.NoError(t, err, "Failed to create enforcer")

	handler := app.NewHasWorkspacePermissionHandler(e)

	tests := []struct {
		name        string
		userID      string
		workspaceID string
		permission  app.WorkspacePermission
		expected    bool
	}{
		{"W111-Owner: Read Workspace", "111", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionRead, true},
		{"W111-Owner: Edit Workspace", "111", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionEdit, true},
		{"W111-Owner: Delete Workspace", "111", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionDelete, true},
		{"W111-Editor: Read Workspace", "112", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionRead, true},
		{"W111-Editor: CANNOT Edit Workspace", "112", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionEdit, false},
		{"W111-Editor: CANNOT Delete Workspace", "112", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionDelete, false},
		{"W111-Viewer: Read Workspace", "110", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionRead, true},
		{"W111-Viewer: CANNOT Edit Workspace", "110", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionEdit, false},
		{"W112-Owner: Delete Workspace", "112", "00000000-0000-4000-8000-000000000112", app.WorkspacePermissionDelete, true},
		{"W112-Editor: Read Workspace", "111", "00000000-0000-4000-8000-000000000112", app.WorkspacePermissionRead, true},
		{"W112-Editor: CANNOT Delete Workspace", "111", "00000000-0000-4000-8000-000000000112", app.WorkspacePermissionDelete, false},
		{"W112-Stranger: CANNOT Read Workspace", "110", "00000000-0000-4000-8000-000000000112", app.WorkspacePermissionRead, false},
		{"W110-Owner: Edit Workspace", "110", "00000000-0000-4000-8000-000000000110", app.WorkspacePermissionEdit, true},
		{"W110-Owner user 111: Edit Workspace", "111", "00000000-0000-4000-8000-000000000110", app.WorkspacePermissionEdit, true},
		{"W110-Owner user 112: Edit Workspace", "112", "00000000-0000-4000-8000-000000000110", app.WorkspacePermissionEdit, true},
		{"Security: user:111 (Owner of W111) cannot edit W112", "111", "00000000-0000-4000-8000-000000000112", app.WorkspacePermissionEdit, false},
		{"Security: user:112 (Owner of W112) cannot delete W111", "112", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionDelete, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			workspaceID := uuid.MustParse(tc.workspaceID)
			ok, err := handler.Handle(ctx, &app.HasWorkspacePermission{
				UserID:      tc.userID,
				WorkspaceID: workspaceID,
				Permission:  tc.permission,
			})
			require.NoError(t, err, "Handler threw an error")
			assert.Equal(t, tc.expected, ok)
		})
	}
}
