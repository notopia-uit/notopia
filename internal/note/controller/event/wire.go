package event

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideIntegration,
	ProvideWatermillLogger,
	ProvideWatermillMarshaler,
)
