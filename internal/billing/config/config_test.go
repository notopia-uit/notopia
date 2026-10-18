package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const requiredConfig = `
server:
  url: http://localhost
database:
  host: localhost
  name: billing
  user: u
  password: p
kafka:
  brokers: [localhost:9092]
redis:
  addr: localhost:6379
identity:
  service_account_user_id: svc
services:
  authorization:
    url: localhost:1
    live_url: http://localhost:2
stripe:
  secret_key: sk
  webhook_secret: whsec
billing:
  web_base_url: http://localhost
`

func writeRequiredConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "billing.notopia.config.yaml"), []byte(requiredConfig), 0o600))
	t.Chdir(dir)
}

func TestCronDefaults(t *testing.T) {
	writeRequiredConfig(t)
	cfg, err := config.New(validator.New(), viper.New())
	require.NoError(t, err)

	assert.Equal(t, "billing:reconcile-subscriptions", cfg.Cron.ReconcileSubscriptions.Name)
	assert.Equal(t, "0 * * * *", cfg.Cron.ReconcileSubscriptions.Schedule)
	assert.Equal(t, 15*time.Minute, cfg.Cron.ReconcileSubscriptions.Timeout)
	assert.True(t, cfg.Cron.ReconcileSubscriptions.Enabled)
	assert.Equal(t, "0 4 * * *", cfg.Cron.PruneProviderEvents.Schedule)
	assert.True(t, cfg.Cron.PruneProviderEvents.Enabled)
}

func TestCronEnvOverrides(t *testing.T) {
	writeRequiredConfig(t)
	t.Setenv("NOTOPIA_BILLING_CRON_RECONCILE_SUBSCRIPTIONS_SCHEDULE", "*/5 * * * *")
	t.Setenv("NOTOPIA_BILLING_CRON_RECONCILE_SUBSCRIPTIONS_TIMEOUT", "2m")
	t.Setenv("NOTOPIA_BILLING_CRON_RECONCILE_SUBSCRIPTIONS_NAME", "billing:custom")
	t.Setenv("NOTOPIA_BILLING_CRON_PRUNE_PROVIDER_EVENTS_ENABLED", "false")
	cfg, err := config.New(validator.New(), viper.New())
	require.NoError(t, err)

	assert.Equal(t, "*/5 * * * *", cfg.Cron.ReconcileSubscriptions.Schedule)
	assert.Equal(t, 2*time.Minute, cfg.Cron.ReconcileSubscriptions.Timeout)
	assert.Equal(t, "billing:custom", cfg.Cron.ReconcileSubscriptions.Name)
	assert.True(t, cfg.Cron.ReconcileSubscriptions.Enabled)
	assert.False(t, cfg.Cron.PruneProviderEvents.Enabled)
}
