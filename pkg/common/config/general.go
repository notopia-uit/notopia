package commonconfig

import (
	"github.com/spf13/viper"
)

type AppEnv string

const (
	AppEnvDevelopment AppEnv = "development"
	AppEnvProduction  AppEnv = "production"
)

type General struct {
	AppEnv AppEnv `json:"app_env" mapstructure:"app_env" validate:"omitempty,oneof=development production" yaml:"app_env"`
	TZ     string `json:"tz"      mapstructure:"tz"      validate:"omitempty"                              yaml:"tz"`
}

func GeneralViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".app_env", "production")
	viper.SetDefault(prefix+".tz", "Asia/HoChiMinh")
}
