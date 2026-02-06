package note

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type NoteHTTPServerConfig struct {
	Host string `json:"host" mapstructure:"host" validate:"hostname"      yaml:"host"`
	Port int    `json:"port" mapstructure:"port" validate:"required,port" yaml:"port"`
}

type NoteGRPCServerConfig struct {
	Host string `json:"host" mapstructure:"host" validate:"required,hostname" yaml:"host"`
	Port int    `json:"port" mapstructure:"port" validate:"required,port"     yaml:"port"`
}

type Config struct {
	HTTP     NoteHTTPServerConfig `json:"http"     mapstructure:"http"     validate:"required" yaml:"http"`
	GRPC     NoteGRPCServerConfig `json:"grpc"     mapstructure:"grpc"     validate:"required" yaml:"grpc"`
	OTLP     config.OTLP          `json:"otlp"     mapstructure:"otlp"     validate:"required" yaml:"otlp"`
	Database config.PG            `json:"database" mapstructure:"database" validate:"required" yaml:"database"`
}

func LoadConfig(validate *validator.Validate) (*Config, error) {
	v := viper.New()

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8081)
	v.SetDefault("grpc.host", "0.0.0.0")
	v.SetDefault("grpc.port", 18091)

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
