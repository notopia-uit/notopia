package pubsub

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideRedisClient,
	ProvideWorkspaceEvent,
)
