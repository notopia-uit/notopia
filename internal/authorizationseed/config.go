package authorizationseed

import (
	"github.com/go-playground/validator/v10"
	authorizationconfig "github.com/notopia-uit/notopia/internal/authorization/config"
	"github.com/spf13/viper"
)

type Config = authorizationconfig.SeedConfig

var (
	NewConfig     = authorizationconfig.NewSeedConfig
	ProvideConfig = NewConfig
)

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
