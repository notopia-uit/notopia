package otel

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideLoggerProvider,
	ProvideMeterProvider,
	ProvideResource,
	ProvideSlogHandler,
	ProvideTracerProvider,
)
