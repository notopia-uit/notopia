package commonconfig

import "github.com/spf13/viper"

type Authentik struct {
	Host  string `json:"host"  mapstructure:"host"  validate:"required" yaml:"host"`
	URL   string `json:"url"   mapstructure:"url"   validate:"required" yaml:"url"`
	Token string `json:"token" mapstructure:"token" validate:"required" yaml:"token"`
}

func AuthentikViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".url", "/api/v3")
}
