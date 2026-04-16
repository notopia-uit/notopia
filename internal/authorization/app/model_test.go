package app_test

import (
	"testing"

	_ "embed"
	"github.com/stretchr/testify/require"
)

// Shared at https://editor.casbin.org/#GURKL5ZXW
func TestModels(t *testing.T) {
	e, err := GetLocalEnforcer(t, true)
	require.NoError(t, err, "Failed to create enforcer")

	tests := []struct {
		name  string
		sub   string
		dom   string
		obj   string
		act   string
		allow bool
	}{
		// =====================================================================
		// WORKSPACE 00000000-0000-0000-0000-000000000111: Owner(111), Editor(112), Viewer(110)
		// =====================================================================

		// Owner: user:111
		{"W111-Owner: Read Workspace", "user:111", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "read", true},
		{"W111-Owner: Edit Workspace", "user:111", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "edit", true},
		{"W111-Owner: Delete Workspace", "user:111", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "delete", true},
		{"W111-Owner: Read Note", "user:111", "workspace:00000000-0000-0000-0000-000000000111", "note", "read", true},
		{"W111-Owner: Write Note", "user:111", "workspace:00000000-0000-0000-0000-000000000111", "note", "write", true},
		{"W111-Owner: Delete Folder", "user:111", "workspace:00000000-0000-0000-0000-000000000111", "folder", "delete", true},

		// Editor: user:112
		{"W111-Editor: Read Workspace", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "read", true},
		{"W111-Editor: CANNOT Edit Workspace", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "edit", false},
		{"W111-Editor: CANNOT Delete Workspace", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "delete", false},
		{"W111-Editor: Read Note", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "note", "read", true},
		{"W111-Editor: Write Note", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "note", "write", true},
		{"W111-Editor: Delete Note", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "note", "delete", true},
		{"W111-Editor: Write Folder", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "folder", "write", true},

		// Viewer: user:110
		{"W111-Viewer: Read Workspace", "user:110", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "read", true},
		{"W111-Viewer: CANNOT Edit Workspace", "user:110", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "edit", false},
		{"W111-Viewer: Read Note", "user:110", "workspace:00000000-0000-0000-0000-000000000111", "note", "read", true},
		{"W111-Viewer: CANNOT Write Note", "user:110", "workspace:00000000-0000-0000-0000-000000000111", "note", "write", false},
		{"W111-Viewer: CANNOT Delete Folder", "user:110", "workspace:00000000-0000-0000-0000-000000000111", "folder", "delete", false},

		// =====================================================================
		// WORKSPACE 00000000-0000-0000-0000-000000000112: Owner(112), Editor(111), No Role(110)
		// =====================================================================

		// Owner: user:112
		{"W112-Owner: Delete Workspace", "user:112", "workspace:00000000-0000-0000-0000-000000000112", "workspace", "delete", true},
		{"W112-Owner: Write Note", "user:112", "workspace:00000000-0000-0000-0000-000000000112", "note", "write", true},

		// Editor: user:111
		{"W112-Editor: Read Workspace", "user:111", "workspace:00000000-0000-0000-0000-000000000112", "workspace", "read", true},
		{"W112-Editor: CANNOT Delete Workspace", "user:111", "workspace:00000000-0000-0000-0000-000000000112", "workspace", "delete", false},
		{"W112-Editor: Write Note", "user:111", "workspace:00000000-0000-0000-0000-000000000112", "note", "write", true},

		// No Role: user:110
		{"W112-Stranger: CANNOT Read Note", "user:110", "workspace:00000000-0000-0000-0000-000000000112", "note", "read", false},
		{"W112-Stranger: CANNOT Read Workspace", "user:110", "workspace:00000000-0000-0000-0000-000000000112", "workspace", "read", false},

		// =====================================================================
		// WORKSPACE 00000000-0000-0000-0000-000000000110: Owner(110)
		// =====================================================================

		{"W110-Owner: Edit Workspace", "user:110", "workspace:00000000-0000-0000-0000-000000000110", "workspace", "edit", true},
		{"W110-Stranger: User 111 CANNOT Read W110", "user:111", "workspace:00000000-0000-0000-0000-000000000110", "note", "read", false},
		{"W110-Stranger: User 112 CANNOT Read W110", "user:112", "workspace:00000000-0000-0000-0000-000000000110", "workspace", "read", false},

		// =====================================================================
		// CROSS-TENANT ATTACK (The "Leaking" Test)
		// =====================================================================

		{"Security: user:111 (Owner of W111) cannot edit W112", "user:111", "workspace:00000000-0000-0000-0000-000000000112", "workspace", "edit", false},
		{"Security: user:112 (Owner of W112) cannot delete W111", "user:112", "workspace:00000000-0000-0000-0000-000000000111", "workspace", "delete", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, ex, err := e.EnforceEx(tc.sub, tc.dom, tc.obj, tc.act)
			require.NoError(t, err, "Enforcer threw an error")
			require.Equal(t, tc.allow, ok, "Details: %v", ex)
		})
	}
}
