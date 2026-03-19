package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateWorkspaceMembersHandler(t *testing.T) {
	t.Skip("UpdateWorkspaceMembersHandler requires a transactional adapter (e.g., GORM), FileAdapter does not support transactions")

	tests := []struct {
		name        string
		requesterID string
		workspaceID string
		newMembers  []app.WorkspaceMember
		expectErr   bool
	}{
		{
			name:        "W111-Owner can update members",
			requesterID: "111",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			newMembers: []app.WorkspaceMember{
				{ID: "111", Role: app.WorkspaceRoleOwner},
				{ID: "112", Role: app.WorkspaceRoleViewer},
			},
			expectErr: false,
		},
		{
			name:        "W111-Editor CANNOT update members",
			requesterID: "112",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			newMembers: []app.WorkspaceMember{
				{ID: "111", Role: app.WorkspaceRoleOwner},
			},
			expectErr: true,
		},
		{
			name:        "W111-Viewer CANNOT update members",
			requesterID: "110",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			newMembers: []app.WorkspaceMember{
				{ID: "111", Role: app.WorkspaceRoleOwner},
			},
			expectErr: true,
		},
		{
			name:        "W112-Stranger CANNOT update members",
			requesterID: "110",
			workspaceID: "00000000-0000-0000-0000-000000000112",
			newMembers: []app.WorkspaceMember{
				{ID: "112", Role: app.WorkspaceRoleOwner},
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e, err := GetLocalEnforcer(true)
			require.NoError(t, err, "Failed to create enforcer")

			handler := app.NewUpdateWorkspaceMembersHandler(e)

			workspaceID := uuid.MustParse(tc.workspaceID)
			ctx := context.Background()
			err = handler.Handle(ctx, app.UpdateWorkspaceMembers{
				UserID:      tc.requesterID,
				WorkspaceID: workspaceID,
				Members:     tc.newMembers,
			})

			if tc.expectErr {
				require.Error(t, err, "Expected error but got none")
				return
			}

			require.NoError(t, err, "Handler threw an error")

			getMembersHandler := app.NewGetWorkspaceMembersHandler(e)
			members, err := getMembersHandler.Handle(app.GetWorkspaceMembers{
				UserID:      tc.requesterID,
				WorkspaceID: workspaceID,
			})
			require.NoError(t, err, "Failed to get members")

			require.Len(t, members, len(tc.newMembers), "Expected %d members after update", len(tc.newMembers))

			for _, expected := range tc.newMembers {
				assert.Contains(t, members, expected, "Expected member not found after update")
			}
		})
	}
}
