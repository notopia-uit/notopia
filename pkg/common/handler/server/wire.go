package server

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideGin,
	ProvideGinSlogHandler,
	ProvideOtelGinHandler,
)
