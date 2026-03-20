package pubsub

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideIntegration,
	ProvideIntegrationMarshaler,
	ProvideRedisClient,
	ProvideWatermillLogger,
	ProvideWorkspaceEvent,
	ProvideWorkspaceEventHubPubSub,
	ProvideWorkspaceEventInternalPubSub,
	wire.Bind(new(app.Integration), new(*Integration)),
	wire.Bind(new(app.WorkspaceEvent), new(*WorkspaceEvent)),
)
