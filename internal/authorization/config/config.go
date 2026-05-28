package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	GRPC   commonconfig.ServerAddress `json:"grpc"   mapstructure:"grpc"   validate:"required" yaml:"grpc"`
	Health commonconfig.ServerAddress `json:"health" mapstructure:"health" validate:"required" yaml:"health"`
}

type Config struct {
	General  commonconfig.General `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log     `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Server   ServerConfig         `json:"server"   mapstructure:"server"   validate:"required"  yaml:"server"`
	Database commonconfig.SQL     `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Kafka    commonconfig.Kafka   `json:"kafka"    mapstructure:"kafka"    validate:"required"  yaml:"kafka"`
}

func NewConfig(
	validate *validator.Validate,
	viper *viper.Viper,
) (*Config, error) {
	viper.SetEnvPrefix("notopia_authorization")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("authorization.notopia.config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", viper.ConfigFileUsed()))
	}

	cfg := &Config{}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("cannot set default config: %w", err)
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from env or config file: %w", err)
	}

	if cfg.Server.GRPC.Port == 0 {
		cfg.Server.GRPC.Port = 18089
	}
	if cfg.Server.Health.Port == 0 {
		cfg.Server.Health.Port = 28089
	}
	if cfg.Kafka.ConsumerGroup == "" {
		cfg.Kafka.ConsumerGroup = "authorization-service"
	}

	slog.Info("configuration", slog.Any("config", cfg))

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("Config validation failed: %w", err)
	}

	return cfg, nil
}

var ProvideConfig = NewConfig

func NewViper() *viper.Viper {
	return viper.NewWithOptions(
		viper.ExperimentalBindStruct(),
	)
}

var ProvideViper = NewViper
