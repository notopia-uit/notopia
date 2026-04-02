package otel

import (
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	MapSlogToGRPCMiddlewareLogger,
	ProvideGlobal,
	ProvideLoggerProvider,
	ProvideMeterProvider,
	ProvideOTELSaramaTracer,
	ProvideResource,
	ProvideSlogHandler,
	ProvideTracerProvider,
	wire.Bind(new(kafka.SaramaTracer), new(*WatermillKafkaTracer)),
)
