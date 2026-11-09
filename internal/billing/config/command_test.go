package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestCommandConfig(t *testing.T) {
	for _, tc := range []struct {
		name, command, db, key string
		valid                  bool
	}{{"migrate database only", "migrate", "postgres://localhost/billing", "", true}, {"migrate missing database", "migrate", "", "", false}, {"seed without webhook", "seed", "postgres://localhost/billing", "sk_test", true}, {"seed missing key", "seed", "postgres://localhost/billing", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.db != "" {
				t.Setenv("NOTOPIA_BILLING_DATABASE_URL", tc.db)
			}
			if tc.key != "" {
				t.Setenv("NOTOPIA_BILLING_STRIPE_SECRET_KEY", tc.key)
			}
			t.Setenv("NOTOPIA_BILLING_IDENTITY_SERVICE_ACCOUNT_USER_ID", "1003")
			t.Setenv("NOTOPIA_BILLING_BILLING_WEB_BASE_URL", "http://localhost")
			v := validator.New(validator.WithRequiredStructEnabled())
			var err error
			if tc.command == "migrate" {
				_, err = NewMigrateConfig(v, NewViper())
			} else {
				_, err = NewSeedConfig(v, NewViper())
			}
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
