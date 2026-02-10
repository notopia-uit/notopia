package grpc

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	Provide,
	ProvideHandler,
	wire.Bind(new(IHandler), new(*Handler)),
)
