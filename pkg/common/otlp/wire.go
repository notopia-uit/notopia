package otlp

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideLoggerProvider,
	ProvideMeterProvider,
	ProvideResource,
	ProvideSlog,
	ProvideTracerProvider,
)
