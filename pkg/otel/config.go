package otel

import (
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/log"
)

type RemoteConfig struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint" validate:"omitempty,hostname_port" yaml:"endpoint"`
	Insecure bool   `json:"insecure" mapstructure:"insecure" validate:""                        yaml:"insecure"`
}

type TraceConfig struct {
	Enabled    bool         `json:"enabled"     mapstructure:"enabled"     validate:""            yaml:"enabled"`
	SampleRate float64      `json:"sample_rate" mapstructure:"sample_rate" validate:"gte=0,lte=1" yaml:"sample_rate"`
	GRPC       RemoteConfig `json:"grpc"        mapstructure:"grpc"        validate:""            yaml:"grpc"`
	Stdout     bool         `json:"stdout"      mapstructure:"stdout"      validate:""            yaml:"stdout"`
}

type LogConfig struct {
	Enabled bool         `json:"enabled" mapstructure:"enabled" validate:""                                      yaml:"enabled"`
	Level   string       `json:"level"   mapstructure:"level"   validate:"omitempty,oneof=debug info warn error" yaml:"level"`
	GRPC    RemoteConfig `json:"grpc"    mapstructure:"grpc"    validate:""                                      yaml:"grpc"`
	Stdout  bool         `json:"stdout"  mapstructure:"stdout"  validate:""                                      yaml:"stdout"`
}

func (lc *LogConfig) GetMinSecurity() log.Severity {
	switch lc.Level {
	case "debug":
		return log.SeverityDebug
	case "warn":
		return log.SeverityWarn
	case "error":
		return log.SeverityError
	default:
		return log.SeverityInfo
	}
}

type MeterConfig struct {
	Enabled  bool          `json:"enabled"         mapstructure:"enabled"         validate:""                      yaml:"enabled"`
	GRPC     RemoteConfig  `json:"grpc"            mapstructure:"grpc"            validate:""                      yaml:"grpc"`
	Stdout   bool          `json:"stdout"          mapstructure:"stdout"          validate:""                      yaml:"stdout"`
	Interval time.Duration `json:"export_interval" mapstructure:"export_interval" validate:"required_with=Enabled" yaml:"export_interval"`
}

type Config struct {
	Enabled bool        `json:"enabled" mapstructure:"enabled" validate:"" yaml:"enabled"`
	Stdout  bool        `json:"stdout"  mapstructure:"stdout"  validate:"" yaml:"stdout"`
	Trace   TraceConfig `json:"trace"   mapstructure:"trace"   validate:"" yaml:"trace"`
	Log     LogConfig   `json:"log"     mapstructure:"log"     validate:"" yaml:"log"`
	Meter   MeterConfig `json:"meter"   mapstructure:"meter"   validate:"" yaml:"meter"`
}

func ViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".log.level", "info")
}
