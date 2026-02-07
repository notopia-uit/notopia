package config_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/notopia-uit/notopia/pkg/common/config"
)

func TestPostGresConfigValidation(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name    string
		cfg     config.SQL
		wantErr bool
	}{
		{
			name: "valid DSN with postgres prefix",
			cfg: config.SQL{
				DSN: "postgres://user:password@localhost:5432/dbname",
			},
			wantErr: false,
		},
		{
			name: "invalid DSN - missing postgres prefix",
			cfg: config.SQL{
				DSN: "mysql://user:password@localhost:5432/dbname",
			},
			wantErr: true,
		},
		{
			name: "invalid DSN - empty string requires Host",
			cfg: config.SQL{
				DSN: "",
			},
			wantErr: true,
		},
		{
			name: "valid individual fields with valid SSLMode",
			cfg: config.SQL{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "require",
			},
			wantErr: false,
		},
		{
			name: "valid individual fields with SSLMode disable",
			cfg: config.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "disable",
			},
			wantErr: false,
		},
		{
			name: "valid individual fields with empty SSLMode",
			cfg: config.SQL{
				Host: "localhost",
				Port: 5432,
			},
			wantErr: false,
		},
		{
			name: "valid with DSN and empty Host/Port",
			cfg: config.SQL{
				DSN:  "postgres://user:password@localhost:5432/dbname",
				Host: "",
				Port: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid SSLMode - not in allowed values",
			cfg: config.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid SSLMode allow",
			cfg: config.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "allow",
			},
			wantErr: false,
		},
		{
			name: "valid SSLMode prefer",
			cfg: config.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "prefer",
			},
			wantErr: false,
		},
		{
			name: "valid SSLMode verify-ca",
			cfg: config.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "verify-ca",
			},
			wantErr: false,
		},
		{
			name: "valid SSLMode verify-full",
			cfg: config.SQL{
				Host:    "localhost",
				Port:    5432,
				SSLMode: "verify-full",
			},
			wantErr: false,
		},
		{
			name: "valid port omitted (zero value skipped by omitempty)",
			cfg: config.SQL{
				Host: "localhost",
				Port: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid port - negative value",
			cfg: config.SQL{
				Host: "localhost",
				Port: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid port - above maximum",
			cfg: config.SQL{
				Host: "localhost",
				Port: 65536,
			},
			wantErr: true,
		},
		{
			name: "valid port at minimum boundary",
			cfg: config.SQL{
				Host: "localhost",
				Port: 1,
			},
			wantErr: false,
		},
		{
			name: "valid port at maximum boundary",
			cfg: config.SQL{
				Host: "localhost",
				Port: 65535,
			},
			wantErr: false,
		},
		{
			name: "invalid hostname format - underscore not allowed",
			cfg: config.SQL{
				Host: "invalid_host_with_underscore",
				Port: 5432,
			},
			wantErr: true,
		},
		{
			name:    "missing required fields - neither DSN nor Host",
			cfg:     config.SQL{},
			wantErr: true,
		},
		{
			name: "missing Host when DSN not provided",
			cfg: config.SQL{
				Port: 5432,
			},
			wantErr: true,
		},
		{
			name: "valid FQDN hostname",
			cfg: config.SQL{
				Host: "db.example.com",
				Port: 5432,
			},
			wantErr: false,
		},
		{
			name: "valid IP address hostname",
			cfg: config.SQL{
				Host: "192.168.1.1",
				Port: 5432,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validation failed: got error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDSN(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.SQL
		expected string
	}{
		{
			name: "returns provided DSN",
			cfg: config.SQL{
				DSN: "postgres://user:password@localhost:5432/dbname",
			},
			expected: "postgres://user:password@localhost:5432/dbname",
		},
		{
			name: "constructs DSN from fields",
			cfg: config.SQL{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "require",
			},
			expected: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetDSN()
			if got != tt.expected {
				t.Errorf("GetDSN() = %q, want %q", got, tt.expected)
			}
		})
	}
}
