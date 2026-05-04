package commonconfig

import (
	"github.com/spf13/viper"
)

type Authentik struct {
	Host   string `json:"host"   mapstructure:"host"   validate:"required,hostname"          yaml:"host"`
	Scheme string `json:"scheme" mapstructure:"scheme" validate:"omitempty,oneof=http https" yaml:"scheme"`
	URL    string `json:"url"    mapstructure:"url"    validate:"required"                   yaml:"url"`
	Token  string `json:"token"  mapstructure:"token"  validate:"required"                   yaml:"token"`
}

func (a *Authentik) HealthLiveURL() string {
	return a.Scheme + "://" + a.Host + "/-/health/live"
}

func AuthentikViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".scheme", "http")
	viper.SetDefault(prefix+".url", "/api/v3")
}
