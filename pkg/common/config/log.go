package commonconfig

import (
	"log/slog"
)

type Log struct {
	Enabled bool   `default:"true" json:"enabled" mapstructure:"enabled" validate:""                                      yaml:"enabled"`
	Level   string `default:"info" json:"level"   mapstructure:"level"   validate:"omitempty,oneof=debug info warn error" yaml:"level"`
}

func (c *Log) GetSlogLevel() slog.Level {
	var level slog.Level
	err := level.UnmarshalText([]byte(c.Level))
	if err != nil {
		panic(err) // this should never happen due to validation, but we panic just in case
	}
	return level
}
