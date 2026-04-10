package grpc

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/pb"
)

var ProviderSet = wire.NewSet(
	ProvideServiceServer,
	ProvideServiceServerAdapter,
	Provide,
	wire.Bind(new(pb.NoteServiceServer), new(*ServiceServer)),
)
