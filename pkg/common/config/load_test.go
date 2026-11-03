package commonconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadPrecedence(t *testing.T) {
	type section struct {
		Value string `default:"default" mapstructure:"value" validate:"required"`
	}
	type cfg struct {
		Section section `mapstructure:"section" validate:"required"`
	}
	for _, tc := range []struct{ name, file, env, want string }{
		{"default", "", "", "default"}, {"file", "section:\n  value: file\n", "", "file"}, {"environment", "section:\n  value: file\n", "environment", "environment"}, {"environment without file", "", "environment", "environment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.file != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "test.yaml"), []byte(tc.file), 0o600))
			}
			if tc.env != "" {
				t.Setenv("TEST_LOAD_SECTION_VALUE", tc.env)
			}
			v := viper.New()
			v.AddConfigPath(dir)
			got, err := Load[cfg](validator.New(validator.WithRequiredStructEnabled()), v, "test_load", "test")
			require.NoError(t, err)
			require.Equal(t, tc.want, got.Section.Value)
		})
	}
}

func TestRedisOptions(t *testing.T) {
	r := Redis{Addr: "localhost:6379", Username: "user", Password: Secret("password"), DB: 3}
	opts := r.Options()
	require.Equal(t, "localhost:6379", opts.Addr)
	require.Equal(t, "user", opts.Username)
	require.Equal(t, "password", opts.Password)
	require.Equal(t, 3, opts.DB)
}
