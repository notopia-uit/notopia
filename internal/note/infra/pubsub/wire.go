package pubsub

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideRedisClient,
	ProvideWorkspaceEventInternalPubSub,
	ProvideWatermillLogger,
	wire.Struct(new(cqrs.JSONMarshaler), "*"),
	wire.Bind(new(cqrs.CommandEventMarshaler), new(*cqrs.JSONMarshaler)),
)
