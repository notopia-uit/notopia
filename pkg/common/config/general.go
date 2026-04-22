package commonconfig

import (
	"github.com/spf13/viper"
)

type General struct {
	TZ string `json:"tz" mapstructure:"tz" validate:"omitempty" yaml:"tz"`
}

func GeneralViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".tz", "Asia/HoChiMinh")
}
