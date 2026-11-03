package billingseed

import (
	"github.com/go-playground/validator/v10"
	billingconfig "github.com/notopia-uit/notopia/internal/billing/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type (
	Stripe  = billingconfig.StripeAPI
	Catalog billingconfig.Billing
)

type Config struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Stripe   Stripe                `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Catalog  Catalog               `json:"catalog"  mapstructure:"catalog"  validate:"required"  yaml:"catalog"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
}

func NewConfig(validate *validator.Validate, v *viper.Viper) (*Config, error) {
	cfg, err := billingconfig.NewSeedConfig(validate, v)
	if err != nil {
		return nil, err
	}
	return &Config{General: cfg.General, Log: cfg.Log, Database: cfg.Database, Stripe: cfg.Stripe, Catalog: Catalog(cfg.Billing), Identity: cfg.Identity}, nil
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
