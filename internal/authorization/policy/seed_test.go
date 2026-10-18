package policy_test

import (
	"path/filepath"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"github.com/notopia-uit/notopia/internal/authorization/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSeed_AddsMissingRowsOnExistingDatabase(t *testing.T) {
	enforcer := newPersistentEnforcer(t, filepath.Join(t.TempDir(), "casbin.db"))
	_, err := enforcer.RemovePolicies([][]string{
		{"owner", "workspace", "billing_read"},
		{"owner", "workspace", "billing_manage"},
		{"editor", "workspace", "billing_read"},
		{"viewer", "workspace", "billing_read"},
	})
	require.NoError(t, err)

	hasBilling, err := enforcer.HasPolicy("owner", "workspace", "billing_read")
	require.NoError(t, err)
	require.False(t, hasBilling)

	require.NoError(t, policy.Seed(enforcer))

	for _, rule := range [][]string{
		{"owner", "workspace", "read"},
		{"owner", "workspace", "billing_read"},
		{"owner", "workspace", "billing_manage"},
		{"editor", "workspace", "billing_read"},
		{"viewer", "workspace", "billing_read"},
	} {
		ok, err := enforcer.HasPolicy(rule)
		require.NoError(t, err)
		assert.True(t, ok, rule)
	}

	require.NoError(t, policy.Seed(enforcer), "seed must be idempotent")

	require.NoError(t, enforcer.LoadPolicy())
	ok, err := enforcer.HasPolicy("viewer", "workspace", "billing_read")
	require.NoError(t, err)
	assert.True(t, ok, "rows must be persisted to the adapter")
}

func newPersistentEnforcer(t *testing.T, path string) *casbin.TransactionalEnforcer {
	t.Helper()

	m, err := model.NewModelFromString(policy.ModelConf)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	adapter, err := gormadapter.NewTransactionalAdapterByDB(db)
	require.NoError(t, err)

	enforcer, err := casbin.NewTransactionalEnforcer(m, adapter)
	require.NoError(t, err)
	return enforcer
}

func TestSeed_KeepsPersistedMembershipsAcrossRestart(t *testing.T) {
	const workspace = "workspace:aaaaaaaa-0000-4000-8000-000000000001"
	path := filepath.Join(t.TempDir(), "casbin.db")

	first := newPersistentEnforcer(t, path)
	require.NoError(t, policy.Seed(first))
	added, err := first.AddGroupingPolicy("user:999", "owner", workspace)
	require.NoError(t, err)
	require.True(t, added)

	restarted := newPersistentEnforcer(t, path)
	require.NoError(t, policy.Seed(restarted))

	ok, err := restarted.Enforce("user:999", workspace, "workspace", "read")
	require.NoError(t, err)
	assert.True(t, ok, "persisted membership must survive seed")

	require.NoError(t, policy.LoadFixtures(restarted))

	ok, err = restarted.Enforce("user:999", workspace, "workspace", "read")
	require.NoError(t, err)
	assert.True(t, ok, "persisted membership must survive fixtures")

	ok, err = restarted.Enforce(
		"user:111",
		"workspace:00000000-0000-4000-8000-000000000111",
		"workspace",
		"read",
	)
	require.NoError(t, err)
	assert.True(t, ok, "fixture membership must be effective")
}

func TestSeed_PrunesStalePolicies(t *testing.T) {
	enforcer := newPersistentEnforcer(t, filepath.Join(t.TempDir(), "casbin.db"))
	_, err := enforcer.AddPolicy("viewer", "workspace", "delete")
	require.NoError(t, err)

	require.NoError(t, policy.Seed(enforcer))
	require.NoError(t, policy.Seed(enforcer))

	stale, err := enforcer.HasPolicy("viewer", "workspace", "delete")
	require.NoError(t, err)
	assert.False(t, stale)

	kept, err := enforcer.HasPolicy("viewer", "workspace", "read")
	require.NoError(t, err)
	assert.True(t, kept)

	require.NoError(t, enforcer.LoadPolicy())
	stale, err = enforcer.HasPolicy("viewer", "workspace", "delete")
	require.NoError(t, err)
	assert.False(t, stale, "prune must be persisted")
}

func TestMissingRules(t *testing.T) {
	enforcer := newPersistentEnforcer(t, filepath.Join(t.TempDir(), "casbin.db"))

	missing, err := policy.MissingRules(enforcer)
	require.NoError(t, err)
	assert.NotEmpty(t, missing)

	require.NoError(t, policy.Seed(enforcer))

	missing, err = policy.MissingRules(enforcer)
	require.NoError(t, err)
	assert.Empty(t, missing)
}
