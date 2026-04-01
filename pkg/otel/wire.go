package otel

import (
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideLoggerProvider,
	ProvideMeterProvider,
	ProvideResource,
	ProvideSlogHandler,
	ProvideTracerProvider,
	ProvideOTELSaramaTracer,
	MapSlogToGRPCMiddlewareLogger,
	wire.Bind(new(kafka.SaramaTracer), new(*WatermillKafkaTracer)),
)
