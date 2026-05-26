package commonconfig

import "time"

type Service struct {
	URL               string        `default:"-"  json:"url"               mapstructure:"url"                validate:"required,hostname_port" yaml:"url"`
	LiveURL           string        `default:"-"  json:"liveUrl"           mapstructure:"live_url"           validate:"required,url"           yaml:"liveUrl"`
	ConnectionTimeout time.Duration `default:"5s" json:"connectionTimeout" mapstructure:"connection_timeout" validate:"required"               yaml:"connectionTimeout"`
}
