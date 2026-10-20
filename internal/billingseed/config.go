package billingseed

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type Stripe struct {
	SecretKey commonconfig.Secret `default:"-" json:"-" mapstructure:"secret_key" validate:"required" yaml:"secret_key"`
}

type Catalog struct {
	Currency   string `default:"vnd" json:"currency"   mapstructure:"currency"     validate:"required"     yaml:"currency"`
	WebBaseURL string `default:"-"   json:"webBaseUrl" mapstructure:"web_base_url" validate:"required,url" yaml:"web_base_url"`
}

type Config struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Stripe   Stripe                `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Catalog  Catalog               `json:"catalog"  mapstructure:"catalog"  validate:"required"  yaml:"catalog"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
}

func NewConfig(validate *validator.Validate, viper *viper.Viper) (*Config, error) {
	viper.SetEnvPrefix("notopia_billingseed")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("billingseed.notopia.config")
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
	slog.Info("configuration", slog.Any("config", cfg))
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
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

func NewValidate() *validator.Validate {
	return validator.New(
		validator.WithRequiredStructEnabled(),
	)
}

var ProvideValidate = NewValidate
