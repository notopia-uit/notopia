package config

import (
	"time"

	"go.opentelemetry.io/otel/log"
)

type OTLPTrace struct {
	Enabled    bool        `json:"enabled"     mapstructure:"enabled"     validate:""             yaml:"enabled"`
	SampleRate float64     `json:"sample_rate" mapstructure:"sample_rate" validate:"gte=0,lte=1"  yaml:"sample_rate"`
	GRPC       *OTLPRemote `json:"grpc"        mapstructure:"grpc"        validate:"omitnil,dive" yaml:"grpc"`
	Stdout     bool        `json:"stdout"      mapstructure:"stdout"      validate:""             yaml:"stdout"`
}

type OTLPRemote struct {
	Endpoint string `json:"endpoint"  mapstructure:"endpoint"  validate:"omitempty,hostname_port" yaml:"endpoint"`
	InSecure bool   `json:"in_secure" mapstructure:"in_secure" validate:"required"                yaml:"in_secure"` // Currently required, because we don't want to handle TLS
}

type OTLPLog struct {
	Enabled bool        `json:"enabled" mapstructure:"enabled" validate:""                                      yaml:"enabled"`
	Level   string      `json:"level"   mapstructure:"level"   validate:"omitempty,oneof=debug info warn error" yaml:"level"`
	GRPC    *OTLPRemote `json:"grpc"    mapstructure:"grpc"    validate:"omitnil,dive"                          yaml:"grpc"`
	Stdout  bool        `json:"stdout"  mapstructure:"stdout"  validate:""                                      yaml:"stdout"`
}

func (o *OTLPLog) GetOTelSeverity() log.Severity {
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
	GRPC     *OTLPRemote   `json:"grpc"            mapstructure:"grpc"            validate:"omitnil,dive"          yaml:"grpc"`
	Stdout   bool          `json:"stdout"          mapstructure:"stdout"          validate:""                      yaml:"stdout"`
	Interval time.Duration `json:"export_interval" mapstructure:"export_interval" validate:"required_with=Enabled" yaml:"export_interval"`
}

type OTLP struct {
	Enabled bool        `json:"enabled" mapstructure:"enabled" validate:""             yaml:"enabled"`
	GRPC    *OTLPRemote `json:"grpc"    mapstructure:"grpc"    validate:"omitnil,dive" yaml:"grpc"`
	Stdout  bool        `json:"stdout"  mapstructure:"stdout"  validate:""             yaml:"stdout"`
	Trace   *OTLPTrace  `json:"trace"   mapstructure:"trace"   validate:"omitnil,dive" yaml:"trace"`
	Log     *OTLPLog    `json:"log"     mapstructure:"log"     validate:"omitnil,dive" yaml:"log"`
	Meter   *OTLPMeter  `json:"meter"   mapstructure:"meter"   validate:"omitnil,dive" yaml:"meter"`
}

func (o *OTLP) TraceEnabled() bool {
	return o != nil &&
		o.Enabled &&
		o.Trace != nil &&
		o.Trace.Enabled
}

func (o *OTLP) TraceEndpoint() (*OTLPRemote, bool) {
	if o.Trace.GRPC != nil && o.Trace.GRPC.Endpoint != "" {
		return o.Trace.GRPC, true
	}
	if o.GRPC != nil {
		return o.GRPC, true
	}
	return nil, false
}

func (o *OTLP) TraceStdoutEnabled() bool {
	return o.Trace.Stdout
}

func (o *OTLP) LogEnabled() bool {
	return o != nil &&
		o.Enabled &&
		o.Log != nil &&
		o.Log.Enabled
}

func (o *OTLP) LogEndpoint() (*OTLPRemote, bool) {
	if o.Log.GRPC != nil && o.Log.GRPC.Endpoint != "" {
		return o.Log.GRPC, true
	}
	if o.GRPC != nil {
		return o.GRPC, true
	}
	return nil, false
}

func (o *OTLP) LogStdoutEnabled() bool {
	return o.Log.Stdout
}

func (o *OTLP) MeterEnabled() bool {
	return o != nil &&
		o.Enabled &&
		o.Meter != nil &&
		o.Meter.Enabled
}

func (o *OTLP) MeterEndpoint() (*OTLPRemote, bool) {
	if o.Meter.GRPC != nil && o.Meter.GRPC.Endpoint != "" {
		return o.Meter.GRPC, true
	}
	if o.GRPC != nil {
		return o.GRPC, true
	}
	return nil, false
}

func (o *OTLP) MeterStdoutEnabled() bool {
	return o.Meter.Stdout
}

func (o *OTLP) MetricExportInterval() time.Duration {
	if o.Meter != nil && o.Meter.Interval > 0 {
		return o.Meter.Interval
	}
	return 30
}
