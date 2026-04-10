package commonconfig

type Authentik struct {
	Host  string `json:"host"  mapstructure:"host"  validate:"required" yaml:"host"`
	URL   string `json:"url"   mapstructure:"url"   validate:"required" yaml:"url"`
	Token string `json:"token" mapstructure:"token" validate:"required" yaml:"token"`
}
