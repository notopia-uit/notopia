package config

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProvideSet = wire.NewSet(
	wire.Bind(new(otel.TraceConfig), new(*OTLPTrace)),
	wire.Bind(new(otel.LogConfig), new(*OTLPLog)),
	wire.Bind(new(otel.MeterConfig), new(*OTLPMeter)),
	wire.FieldsOf(new(*OTLP), "Log", "Trace", "Meter"),
)
