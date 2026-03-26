package app_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkspaceMembersHandler(t *testing.T) {
	e, err := GetLocalEnforcer(true)
	require.NoError(t, err, "Failed to create enforcer")

	handler := app.NewGetWorkspaceMembersHandler(e)

	tests := []struct {
		name            string
		requesterID     string
		workspaceID     string
		expectErr       bool
		expectedMembers []app.WorkspaceMember
	}{
		{
			name:        "W111-Owner can view members",
			requesterID: "111",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			expectErr:   false,
			expectedMembers: []app.WorkspaceMember{
				{ID: "111", Role: app.WorkspaceRoleOwner},
				{ID: "112", Role: app.WorkspaceRoleEditor},
				{ID: "110", Role: app.WorkspaceRoleViewer},
			},
		},
		{
			name:        "W111-Editor can view members",
			requesterID: "112",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			expectErr:   false,
			expectedMembers: []app.WorkspaceMember{
				{ID: "111", Role: app.WorkspaceRoleOwner},
				{ID: "112", Role: app.WorkspaceRoleEditor},
				{ID: "110", Role: app.WorkspaceRoleViewer},
			},
		},
		{
			name:        "W111-Viewer can view members",
			requesterID: "110",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			expectErr:   false,
			expectedMembers: []app.WorkspaceMember{
				{ID: "111", Role: app.WorkspaceRoleOwner},
				{ID: "112", Role: app.WorkspaceRoleEditor},
				{ID: "110", Role: app.WorkspaceRoleViewer},
			},
		},
		{
			name:            "W112-Stranger CANNOT view members",
			requesterID:     "110",
			workspaceID:     "00000000-0000-0000-0000-000000000112",
			expectErr:       true,
			expectedMembers: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			workspaceID := uuid.MustParse(tc.workspaceID)
			members, err := handler.Handle(app.GetWorkspaceMembers{
				UserID:      tc.requesterID,
				WorkspaceID: workspaceID,
			})

			if tc.expectErr {
				require.Error(t, err, "Expected error but got none")
				return
			}

			require.NoError(t, err, "Handler threw an error")
			require.Len(t, members, len(tc.expectedMembers))

			for _, expected := range tc.expectedMembers {
				assert.Contains(t, members, expected, "Expected member not found in result")
			}
		})
	}
}
