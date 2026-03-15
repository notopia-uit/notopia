package event

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideWatermillLogger,
	ProvideIntegration,
	ProvideInternal,
)
