package config

import (
	"time"

	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/log"
)

type OTLPRemote struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint" validate:"omitempty,hostname_port" yaml:"endpoint"`
	Insecure bool   `json:"insecure" mapstructure:"insecure" validate:"required"                yaml:"insecure"` // Currently required, because we don't want to handle TLS
}

type OTLPTrace struct {
	Enabled    bool        `json:"enabled"     mapstructure:"enabled"     validate:""            yaml:"enabled"`
	SampleRate float64     `json:"sample_rate" mapstructure:"sample_rate" validate:"gte=0,lte=1" yaml:"sample_rate"`
	GRPC       *OTLPRemote `json:"grpc"        mapstructure:"grpc"        validate:"omitnil"     yaml:"grpc"`
	Stdout     bool        `json:"stdout"      mapstructure:"stdout"      validate:""            yaml:"stdout"`
}

var _ otel.TraceConfig = (*OTLPTrace)(nil)

func (o *OTLPTrace) ShouldExportStdout() bool {
	return o.Stdout
}

func (o *OTLPTrace) ShouldExportGRPC() bool {
	return o.GRPC.Endpoint != ""
}

func (o *OTLPTrace) GetGRPCRemote() *otel.Remote {
	return &otel.Remote{
		Endpoint: o.GRPC.Endpoint,
		Insecure: o.GRPC.Insecure,
	}
}

func (o *OTLPTrace) GetSampleRate() float64 {
	return o.SampleRate
}

type OTLPLog struct {
	Enabled bool        `json:"enabled" mapstructure:"enabled" validate:""                                      yaml:"enabled"`
	Level   string      `json:"level"   mapstructure:"level"   validate:"omitempty,oneof=debug info warn error" yaml:"level"`
	GRPC    *OTLPRemote `json:"grpc"    mapstructure:"grpc"    validate:"omitnil"                               yaml:"grpc"`
	Stdout  bool        `json:"stdout"  mapstructure:"stdout"  validate:""                                      yaml:"stdout"`
}

var _ otel.LogConfig = (*OTLPLog)(nil)

func (o *OTLPLog) ShouldExportStdout() bool {
	return o.Stdout
}

func (o *OTLPLog) ShouldExportGRPC() bool {
	return o.GRPC.Endpoint != ""
}

func (o *OTLPLog) GetGRPCRemote() *otel.Remote {
	return &otel.Remote{
		Endpoint: o.GRPC.Endpoint,
		Insecure: o.GRPC.Insecure,
	}
}

func (o *OTLPLog) GetMinSecurity() log.Severity {
	switch o.Level {
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

type OTLPMeter struct {
	Enabled  bool          `json:"enabled"         mapstructure:"enabled"         validate:""                      yaml:"enabled"`
	GRPC     *OTLPRemote   `json:"grpc"            mapstructure:"grpc"            validate:"omitnil"               yaml:"grpc"`
	Stdout   bool          `json:"stdout"          mapstructure:"stdout"          validate:""                      yaml:"stdout"`
	Interval time.Duration `json:"export_interval" mapstructure:"export_interval" validate:"required_with=Enabled" yaml:"export_interval"`
}

var _ otel.MeterConfig = (*OTLPMeter)(nil)

func (o *OTLPMeter) ShouldExportStdout() bool {
	return o.Stdout
}

func (o *OTLPMeter) ShouldExportGRPC() bool {
	return o.GRPC.Endpoint != ""
}

func (o *OTLPMeter) GetGRPCRemote() *otel.Remote {
	return &otel.Remote{
		Endpoint: o.GRPC.Endpoint,
		Insecure: o.GRPC.Insecure,
	}
}

func (o *OTLPMeter) GetExportInterval() time.Duration {
	return o.Interval
}

type OTLP struct {
	Enabled bool       `json:"enabled" mapstructure:"enabled" validate:""        yaml:"enabled"`
	Stdout  bool       `json:"stdout"  mapstructure:"stdout"  validate:""        yaml:"stdout"`
	Trace   *OTLPTrace `json:"trace"   mapstructure:"trace"   validate:"omitnil" yaml:"trace"`
	Log     *OTLPLog   `json:"log"     mapstructure:"log"     validate:"omitnil" yaml:"log"`
	Meter   *OTLPMeter `json:"meter"   mapstructure:"meter"   validate:"omitnil" yaml:"meter"`
}

func OTLPViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".enabled", true)
	viper.SetDefault(prefix+".log.enabled", true)
	viper.SetDefault(prefix+".log.level", "info")
	viper.SetDefault(prefix+".log.stdout", true)
}
