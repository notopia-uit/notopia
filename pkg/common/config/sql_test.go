package commonconfig_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostGresConfigValidation(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name    string
		cfg     commonconfig.SQL
		wantErr bool
	}{
		{
			name: "valid URL with postgres prefix",
			cfg: commonconfig.SQL{
				URL: "postgres://user:password@localhost:5432/dbname",
			},
			wantErr: false,
		},
		{
			name: "invalid URL - missing postgres prefix",
			cfg: commonconfig.SQL{
				URL: "mysql://user:password@localhost:5432/dbname",
			},
			wantErr: true,
		},
		{
			name: "invalid URL - empty string requires Host",
			cfg: commonconfig.SQL{
				URL: "",
			},
			wantErr: true,
		},
		{
			name: "valid individual fields with valid SSLMode",
			cfg: commonconfig.SQL{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "require",
			},
			wantErr: false,
		},
		{
			name: "valid individual fields with SSLMode disable",
			cfg: commonconfig.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "disable",
			},
			wantErr: false,
		},
		{
			name: "valid individual fields with empty SSLMode",
			cfg: commonconfig.SQL{
				Host: "localhost",
				Port: 5432,
			},
			wantErr: false,
		},
		{
			name: "valid with URL and empty Host/Port",
			cfg: commonconfig.SQL{
				URL:  "postgres://user:password@localhost:5432/dbname",
				Host: "",
				Port: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid SSLMode - not in allowed values",
			cfg: commonconfig.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid SSLMode allow",
			cfg: commonconfig.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "allow",
			},
			wantErr: false,
		},
		{
			name: "valid SSLMode prefer",
			cfg: commonconfig.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "prefer",
			},
			wantErr: false,
		},
		{
			name: "valid SSLMode verify-ca",
			cfg: commonconfig.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "verify-ca",
			},
			wantErr: false,
		},
		{
			name: "valid SSLMode verify-full",
			cfg: commonconfig.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "verify-full",
			},
			wantErr: false,
		},
		{
			name: "valid port omitted (zero value skipped by omitempty)",
			cfg: commonconfig.SQL{
				Host: "localhost",
				Port: 0,
			},
			wantErr: false,
		},
		{
			name: "valid port at minimum boundary",
			cfg: commonconfig.SQL{
				Host: "localhost",
				Port: 1,
			},
			wantErr: false,
		},
		{
			name: "valid port at maximum boundary",
			cfg: commonconfig.SQL{
				Host: "localhost",
				Port: 65535,
			},
			wantErr: false,
		},
		{
			name: "invalid hostname format - underscore not allowed",
			cfg: commonconfig.SQL{
				Host: "invalid_host_with_underscore",
				Port: 5432,
			},
			wantErr: true,
		},
		{
			name:    "missing required fields - neither URL nor Host",
			cfg:     commonconfig.SQL{},
			wantErr: true,
		},
		{
			name: "missing Host when URL not provided",
			cfg: commonconfig.SQL{
				Port: 5432,
			},
			wantErr: true,
		},
		{
			name: "valid FQDN hostname",
			cfg: commonconfig.SQL{
				Host: "db.example.com",
				Port: 5432,
			},
			wantErr: false,
		},
		{
			name: "valid IP address hostname",
			cfg: commonconfig.SQL{
				Host: "192.168.1.1",
				Port: 5432,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      commonconfig.SQL
		expected string
	}{
		{
			name: "returns provided URL",
			cfg: commonconfig.SQL{
				URL: "postgres://user:password@localhost:5432/dbname",
			},
			expected: "postgres://user:password@localhost:5432/dbname",
		},
		{
			name: "constructs URL from fields",
			cfg: commonconfig.SQL{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "require",
			},
			expected: "postgres://testuser:testpass@localhost:5432/testdb?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cfg.GetURL())
		})
	}
}

func TestGetDSN(t *testing.T) {
	cfg := commonconfig.SQL{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
		SSLMode:  "require",
	}
	want := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=require"
	assert.Equal(t, want, cfg.GetDSN())
}
