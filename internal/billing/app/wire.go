package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCmds,
	ProvideQueries,
	wire.Struct(new(Server), "*"),
)
