package app_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserWorkspaceItemPermissionsHandler(t *testing.T) {
	e, err := GetLocalEnforcer(t, &GetLocalEnforcerParams{LoadTestPolicies: true})
	require.NoError(t, err, "Failed to create enforcer")

	handler := app.NewGetUserWorkspaceItemPermissionsHandler(e)

	tests := []struct {
		name        string
		userID      string
		workspaceID string
		expected    app.WorkspaceItemPermissions
		expectErr   bool
	}{
		{"W111-Owner: All permissions", "111", "00000000-0000-4000-8000-000000000111", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
		{"W111-Editor: Read, Write, Delete", "112", "00000000-0000-4000-8000-000000000111", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
		{"W111-Viewer: Read only", "110", "00000000-0000-4000-8000-000000000111", app.WorkspaceItemPermissions{Read: true, Write: false, Delete: false}, false},
		{"W112-Owner: All permissions", "112", "00000000-0000-4000-8000-000000000112", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
		{"W112-Editor: Read, Write, Delete", "111", "00000000-0000-4000-8000-000000000112", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
		{"W112-Stranger: No permissions", "110", "00000000-0000-4000-8000-000000000112", app.WorkspaceItemPermissions{Read: false, Write: false, Delete: false}, true},
		{"W110-Owner: All permissions", "110", "00000000-0000-4000-8000-000000000110", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
		{"W110-Owner user 112: All permissions", "112", "00000000-0000-4000-8000-000000000110", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
		{"W110-Owner user 111: All permissions", "111", "00000000-0000-4000-8000-000000000110", app.WorkspaceItemPermissions{Read: true, Write: true, Delete: true}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			workspaceID := uuid.MustParse(tc.workspaceID)
			perms, err := handler.Handle(ctx, &app.GetUserWorkspaceItemPermissions{
				UserID:      tc.userID,
				WorkspaceID: workspaceID,
			})
			if tc.expectErr {
				require.Error(t, err, "Expected error but got none")
				return
			}
			require.NoError(t, err, "Handler threw an error")
			assert.Equal(t, tc.expected, perms)
		})
	}
}
