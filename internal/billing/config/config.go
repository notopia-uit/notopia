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

type Server struct {
	URL    string                     `json:"url"    mapstructure:"url"    validate:"required,url" yaml:"url"`
	Health commonconfig.ServerAddress `json:"health" mapstructure:"health" validate:"required"     yaml:"health"`
	HTTP   commonconfig.ServerAddress `json:"http"   mapstructure:"http"   validate:"required"     yaml:"http"`
	GRPC   commonconfig.ServerAddress `json:"grpc"   mapstructure:"grpc"   validate:"required"     yaml:"grpc"`
}

type Services struct {
	Authorization commonconfig.Service `json:"authorization" mapstructure:"authorization" validate:"required" yaml:"authorization"`
}

type Stripe struct {
	SecretKey     commonconfig.Secret `default:"-" json:"-" mapstructure:"secret_key"     validate:"required" yaml:"secret_key"`
	WebhookSecret commonconfig.Secret `default:"-" json:"-" mapstructure:"webhook_secret" validate:"required" yaml:"webhook_secret"`
}

type Billing struct {
	Currency   string `default:"vnd" json:"currency"   mapstructure:"currency"     validate:"required"     yaml:"currency"`
	WebBaseURL string `default:"-"   json:"webBaseUrl" mapstructure:"web_base_url" validate:"required,url" yaml:"web_base_url"`
}

type Config struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Server   Server                `json:"server"   mapstructure:"server"   validate:"required"  yaml:"server"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Kafka    commonconfig.Kafka    `json:"kafka"    mapstructure:"kafka"    validate:"required"  yaml:"kafka"`
	Redis    commonconfig.Redis    `json:"redis"    mapstructure:"redis"    validate:"required"  yaml:"redis"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
	Services Services              `json:"services" mapstructure:"services" validate:"required"  yaml:"services"`
	Stripe   Stripe                `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Billing  Billing               `json:"billing"  mapstructure:"billing"  validate:"required"  yaml:"billing"`
}

func New(validate *validator.Validate, viper *viper.Viper) (*Config, error) {
	viper.SetEnvPrefix("notopia_billing")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("billing.notopia.config")
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
	if cfg.Server.HTTP.Port == 0 {
		cfg.Server.HTTP.Port = 8084
	}
	if cfg.Server.GRPC.Port == 0 {
		cfg.Server.GRPC.Port = 18084
	}
	if cfg.Server.Health.Port == 0 {
		cfg.Server.Health.Port = 28084
	}
	if cfg.Kafka.ConsumerGroup == "" {
		cfg.Kafka.ConsumerGroup = "billing-service"
	}
	slog.Info("configuration", slog.Any("config", cfg))
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

var Provide = New
