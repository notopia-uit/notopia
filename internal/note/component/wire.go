package components

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideValidate,
	ProvideWatermillJsonMarshaler,
	ProvideWatermillLogger,
)
