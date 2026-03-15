package pubsub

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideRedisClient,
	ProvideWorkspaceEvent,
	ProvideWatermillLogger,
	wire.Struct(new(cqrs.JSONMarshaler), "*"),
	wire.Bind(new(cqrs.CommandEventMarshaler), new(*cqrs.JSONMarshaler)),
)
