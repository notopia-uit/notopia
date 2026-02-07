package rpc

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHTTPServiceHandler,
	ProvideRPCHandler,
)
