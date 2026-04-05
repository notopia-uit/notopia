package common

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideWatermillJsonMarshaler,
	ProvideWatermillLogger,
)
