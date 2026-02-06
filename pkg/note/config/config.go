package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type Server struct {
	Host string `json:"host" mapstructure:"host" validate:"omitempty,hostname" yaml:"host"`
	Port string `json:"port" mapstructure:"port" validate:"required,port"      yaml:"port"`
}

func (s *Server) Address() string {
	return s.Host + ":" + s.Port
}

type Config struct {
	Server   *Server      `json:"server"   mapstructure:"server"   validate:"required" yaml:"server"`
	OTLP     *config.OTLP `json:"otlp"     mapstructure:"otlp"     validate:"required" yaml:"otlp"`
	Database *config.SQL  `json:"database" mapstructure:"database" validate:"required" yaml:"database"`
}

func ProvideConfig(validate *validator.Validate, v *viper.Viper) (*Config, error) {
	v.SetDefault("server.port", 8081)

	v.SetEnvPrefix("NOTOPIA_NOTE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
