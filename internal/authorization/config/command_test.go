package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestSeedConfig(t *testing.T) {
	for _, tc := range []struct {
		name, db string
		valid    bool
	}{{"database only", "postgres://localhost/authorization", true}, {"missing database", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.db != "" {
				t.Setenv("NOTOPIA_AUTHORIZATION_DATABASE_URL", tc.db)
			}
			_, err := NewSeedConfig(validator.New(validator.WithRequiredStructEnabled()), NewViper())
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
