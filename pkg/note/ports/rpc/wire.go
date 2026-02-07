package rpc

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

var ProviderSet = wire.NewSet(
	ProvideHTTPServiceHandler,
	ProvideRPCHandler,
	wire.Bind(new(pbconnect.NoteServiceHandler), new(*RPCHandler)),
)
