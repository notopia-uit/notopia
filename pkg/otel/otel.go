package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Global any

func ProvideGlobal(
	loggerProvider log.LoggerProvider,
	meterProvider metric.MeterProvider,
	traceProvider trace.TracerProvider,
) Global {
	global.SetLoggerProvider(loggerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(traceProvider)
	return nil
}
