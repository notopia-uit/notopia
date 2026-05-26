package commonconfig

import (
	"time"
)

type Authentik struct {
	Host              string        `default:"-"       json:"host"               mapstructure:"host"               validate:"required,hostname"          yaml:"host"`
	Scheme            string        `default:"http"    json:"scheme"             mapstructure:"scheme"             validate:"omitempty,oneof=http https" yaml:"scheme"`
	URL               string        `default:"/api/v3" json:"url"                mapstructure:"url"                validate:"required"                   yaml:"url"`
	Token             string        `default:"-"       json:"token"              mapstructure:"token"              validate:"required"                   yaml:"token"`
	ConnectionTimeout time.Duration `default:"5s"      json:"connection_timeout" mapstructure:"connection_timeout" validate:"required"                   yaml:"connection_timeout"`
}

func (a *Authentik) HealthLiveURL() string {
	return a.Scheme + "://" + a.Host + "/-/health/live"
}
