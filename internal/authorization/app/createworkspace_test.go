package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkspaceHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	e, err := GetLocalEnforcer(false)
	require.NoError(t, err, "Failed to create enforcer")

	handler := app.NewCreateWorkspaceHandler(e)

	workspaceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	userID := "000"

	err = handler.Handle(ctx, app.CreateWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	require.NoError(t, err, "Failed to create workspace")

	hasOwnerRole, err := e.HasGroupingPolicy("user:000", "owner", "workspace:00000000-0000-0000-0000-000000000000")
	require.NoError(t, err, "Failed to check grouping policy")
	require.True(t, hasOwnerRole, "Expected user:000 to have owner role on workspace:00000000-0000-0000-0000-000000000000")
}
