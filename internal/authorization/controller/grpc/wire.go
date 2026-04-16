package grpc

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	ProvideService,
)
