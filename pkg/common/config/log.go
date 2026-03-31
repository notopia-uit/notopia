package commonconfig

import (
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/helper"
	"github.com/spf13/viper"
)

type Log struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled" validate:""                                      yaml:"enabled"`
	Level   string `json:"level"   mapstructure:"level"   validate:"omitempty,oneof=debug info warn error" yaml:"level"`
}

func (c *Log) GetSlogLevel() slog.Level {
	level, err := helper.GetLogLevelFromString(c.Level)
	if err != nil {
		panic(err) // this should never happen due to validation, but we panic just in case
	}
	return level
}

func LogViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".level", true)
	viper.SetDefault(prefix+".level", "info")
}
