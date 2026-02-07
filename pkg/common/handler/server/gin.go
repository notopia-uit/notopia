package server

import (
	"log/slog"

	ginslog "github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/common/otlp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type GinSlogHandlerFunc gin.HandlerFunc

func NewGinSlogHandler(
	logger *slog.Logger,
) GinSlogHandlerFunc {
	return GinSlogHandlerFunc(ginslog.SetLogger(
		ginslog.WithLogger(
			func(c *gin.Context, _ *slog.Logger) *slog.Logger {
				return logger.With("user_id", c.GetString("X-Forwarded-ID"))
			},
		),
		ginslog.WithSkipPath([]string{
			"/healthz",
			"/readyz",
			"/metrics",
			"/livez",
		}),
	))
}

var ProvideGinSlogHandler = NewGinSlogHandler

type OtelGinHandlerFunc gin.HandlerFunc

func NewOtelGinHandler(
	serviceName otlp.ServiceName,
	meterProvider *metric.MeterProvider,
	traceProvider *trace.TracerProvider,
) OtelGinHandlerFunc {
	return OtelGinHandlerFunc(otelgin.Middleware(
		serviceName.String(),
		otelgin.WithMeterProvider(meterProvider),
		otelgin.WithTracerProvider(traceProvider),
		otelgin.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	))
}

var ProvideOtelGinHandler = NewOtelGinHandler

func NewGin(
	serviceName otlp.ServiceName,
	meterProvider *metric.MeterProvider,
	traceProvider *trace.TracerProvider,
	slogHandler GinSlogHandlerFunc,
	otelHandler OtelGinHandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.Use(gin.HandlerFunc(slogHandler))
	r.Use(gin.HandlerFunc(otelHandler))
	return r
}

var ProvideGin = NewGin
