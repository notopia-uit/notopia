package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type Server struct {
	HTTP *config.ServerAddress `json:"http" mapstructure:"http" validate:"required" yaml:"http"`
	GRPC *config.ServerAddress `json:"grpc" mapstructure:"grpc" validate:"required" yaml:"grpc"`
}

type Config struct {
	Server   *Server      `json:"server"   mapstructure:"server"   validate:"required" yaml:"server"`
	OTLP     *config.OTLP `json:"otlp"     mapstructure:"otlp"     validate:"omitnil"  yaml:"otlp"`
	Database *config.SQL  `json:"database" mapstructure:"database" validate:"omitnil"  yaml:"database"`
}

func NewConfig(
	validate *validator.Validate,
	v *viper.Viper,
) (*Config, error) {
	v.SetEnvPrefix("NOTOPIA_NOTE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigName("note.notopia.config")
	v.AddConfigPath(".")

	v.SetDefault("server.http.port", 8081)
	v.SetDefault("server.grpc.port", 18081)
	config.OTLPViperSetDefault(v, "otlp")
	config.SQLViperSetDefault(v, "database")

	v.AutomaticEnv()
	if err := v.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", v.ConfigFileUsed()))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnmarshalConfig, err)
	}

	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidateConfig, err)
	}

	return &cfg, nil
}

var ProvideConfig = NewConfig
