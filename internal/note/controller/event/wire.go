package event

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideIntegration,
	ProvideInternal,
	ProvideWatermillLogger,
	ProvideWatermillMarshaler,
)
