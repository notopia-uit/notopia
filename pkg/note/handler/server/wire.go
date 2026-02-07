package server

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/handler/server"
)

var ProvideSet = wire.NewSet(
	ProvideHTTPHandler,
	ProvideRPCHTTPHandlerRegister,
	ProvideRPCHandler,
	ProvideServer,
	ProvideStrictHTTPHandler,
	server.ProviderSet,
	wire.Bind(new(IRCPHandler), new(*RPCHandler)),
	wire.Bind(new(IStrictHTTPHandler), new(*StrictHTTPHandler)),
)
