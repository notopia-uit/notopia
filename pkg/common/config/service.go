package commonconfig

type Service struct {
	URL string `json:"url" mapstructure:"url" validate:"required,hostname_port" yaml:"url"`
}
