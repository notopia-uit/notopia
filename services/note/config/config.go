package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/spf13/viper"
)

type Server struct {
	HTTP commonconfig.ServerAddress `json:"http" mapstructure:"http" validate:"required" yaml:"http"`
	GRPC commonconfig.ServerAddress `json:"grpc" mapstructure:"grpc" validate:"required" yaml:"grpc"`
}

type Config struct {
	Server   Server               `json:"server"   mapstructure:"server"   validate:"required"  yaml:"server"`
	OTLP     otel.Config          `json:"otlp"     mapstructure:"otlp"     validate:"omitempty" yaml:"otlp"`
	Database commonconfig.SQL     `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	General  commonconfig.General `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
}

func New(
	validate *validator.Validate,
	viper *viper.Viper,
) (*Config, error) {
	viper.SetEnvPrefix("notopia_note")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("note.notopia.config")
	viper.AddConfigPath(".")

	viper.SetDefault("server.http.port", 8081)
	viper.SetDefault("server.grpc.port", 18081)
	otel.ViperSetDefault(viper, "otlp")
	commonconfig.SQLViperSetDefault(viper, "database")
	commonconfig.GeneralViperSetDefault(viper, "general")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", viper.ConfigFileUsed()))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("Cannot unmarshal config from env or config file: %w", err)
	}

	slog.Info("configuration", slog.Any("config", cfg))

	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("Config validation failed: %w", err)
	}

	return &cfg, nil
}

var Provide = New
