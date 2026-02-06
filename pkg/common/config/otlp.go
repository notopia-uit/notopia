package config

type OTLP struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint" validate:"omitempty,hostname_port" yaml:"endpoint"`
}
