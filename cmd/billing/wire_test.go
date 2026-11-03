package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitializeMigratorNeedsOnlyDatabaseConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("NOTOPIA_BILLING_DATABASE_URL", "postgres://user:pass@localhost:5432/billing?sslmode=disable")
	migrator, cleanup, err := InitializeMigrator(context.Background())
	require.NoError(t, err)
	require.NotNil(t, migrator)
	require.NotNil(t, cleanup)
	cleanup()
}
