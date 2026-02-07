package http

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/controller/http"
)

var ProviderSet = wire.NewSet(
	ProvideHTTPHandler,
	ProvideRPCHTTPHandlerRegister,
	ProvideRPCHandler,
	ProvideServer,
	ProvideStrictHTTPHandler,
	http.ProviderSet,
	wire.Bind(new(IRCPHandler), new(*RPCHandler)),
	wire.Bind(new(IStrictHTTPHandler), new(*StrictHTTPHandler)),
)
