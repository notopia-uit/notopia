package app_test

import (
	"context"
	"testing"

	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootStrapPolicies_AddsMissingRowsOnExistingDatabase(t *testing.T) {
	enforcer, err := GetLocalEnforcer(t, &GetLocalEnforcerParams{UseTransaction: true})
	require.NoError(t, err)

	_, err = enforcer.RemovePolicies([][]string{
		{"owner", "workspace", "billing_read"},
		{"owner", "workspace", "billing_manage"},
		{"editor", "workspace", "billing_read"},
		{"viewer", "workspace", "billing_read"},
	})
	require.NoError(t, err)

	hasBilling, err := enforcer.HasPolicy("owner", "workspace", "billing_read")
	require.NoError(t, err)
	require.False(t, hasBilling)

	a := &app.App{Enforcer: enforcer}
	require.NoError(t, a.BootStrapPolicies(context.Background()))

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

	require.NoError(t, a.BootStrapPolicies(context.Background()), "bootstrap must be idempotent")

	require.NoError(t, enforcer.LoadPolicy())
	ok, err := enforcer.HasPolicy("viewer", "workspace", "billing_read")
	require.NoError(t, err)
	assert.True(t, ok, "rows must be persisted to the adapter")
}
