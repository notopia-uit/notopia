package config

import (
	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	GRPC   commonconfig.ServerAddress `json:"grpc"   mapstructure:"grpc"   validate:"required" yaml:"grpc"`
	Health commonconfig.ServerAddress `json:"health" mapstructure:"health" validate:"required" yaml:"health"`
}

type ServeConfig struct {
	General  commonconfig.General `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log     `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Server   ServerConfig         `json:"server"   mapstructure:"server"   validate:"required"  yaml:"server"`
	Database commonconfig.SQL     `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Kafka    commonconfig.Kafka   `json:"kafka"    mapstructure:"kafka"    validate:"required"  yaml:"kafka"`
}

type Config = ServeConfig

type SeedConfig struct {
	General  commonconfig.General `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log     `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Database commonconfig.SQL     `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
}

func NewSeedConfig(validate *validator.Validate, v *viper.Viper) (*SeedConfig, error) {
	return commonconfig.Load[SeedConfig](validate, v, "notopia_authorization", "authorization.notopia.config")
}

func NewConfig(validate *validator.Validate, v *viper.Viper) (*Config, error) {
	v.SetDefault("server.grpc.port", 18089)
	v.SetDefault("server.health.port", 28089)
	v.SetDefault("kafka.consumer_group", "authorization-service")
	return commonconfig.Load[ServeConfig](validate, v, "notopia_authorization", "authorization.notopia.config")
}

var (
	ProvideConfig     = NewConfig
	ProvideSeedConfig = NewSeedConfig
)

func NewViper() *viper.Viper {
	return viper.NewWithOptions(
		viper.ExperimentalBindStruct(),
	)
}

var ProvideViper = NewViper
