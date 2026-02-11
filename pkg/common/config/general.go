package commonconfig

import (
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/helper"
	"github.com/spf13/viper"
)

type General struct {
	TZ          string `json:"tz"           mapstructure:"tz"           validate:""                                      yaml:"tz"`
	LogLevel    string `json:"log_level"    mapstructure:"log_level"    validate:"omitempty,oneof=debug info warn error" yaml:"log_level"`
	LogDisabled bool   `json:"log_disabled" mapstructure:"log_disabled" validate:""                                      yaml:"log_disabled"`
}

func (g *General) GetSlogLevel() slog.Level {
	level, err := helper.GetLogLevelFromString(g.LogLevel)
	if err != nil {
		panic(err) // this should never happen due to validation, but we panic just in case
	}
	return level
}

func GeneralViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".tz", "Asia/HoChiMinh")
	viper.SetDefault(prefix+".log_level", "info")
}
