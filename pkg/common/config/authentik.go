package commonconfig

import (
	"net/url"

	"github.com/spf13/viper"
)

type Authentik struct {
	Host  string `json:"host"  mapstructure:"host"  validate:"required" yaml:"host"`
	URL   string `json:"url"   mapstructure:"url"   validate:"required" yaml:"url"`
	Token string `json:"token" mapstructure:"token" validate:"required" yaml:"token"`

	hostURL       *url.URL
	healthLiveURL string
}

func (a *Authentik) HealthLiveURL() string {
	if a.hostURL == nil {
		var err error
		a.hostURL, err = url.Parse(a.Host)
		if err != nil {
			panic("invalid authentik host URL: " + err.Error())
		}
		a.healthLiveURL = a.hostURL.ResolveReference(&url.URL{Path: "/-/health/live"}).String()
	}
	return a.healthLiveURL
}

func AuthentikViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".url", "/api/v3")
}
