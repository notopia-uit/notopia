package commonconfig

import (
	"time"
)

type Meilisearch struct {
	// This include scheme before
	Host              string        `default:"-"  json:"host"               mapstructure:"host"               validate:"required,url" yaml:"host"`
	APIKey            string        `default:"-"  json:"api_key"            mapstructure:"api_key"            validate:"required"     yaml:"api_key"`
	ConnectionTimeout time.Duration `default:"5s" json:"connection_timeout" mapstructure:"connection_timeout" validate:"required"     yaml:"connection_timeout"`
}
