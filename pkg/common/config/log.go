package config

type LogConfig struct {
	Level string `json:"level" mapstructure:"level" validate:"required,oneof=debug info warn error dpanic panic fatal" yaml:"level"`
}
