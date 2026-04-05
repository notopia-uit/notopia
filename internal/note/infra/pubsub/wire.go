package pubsub

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationMarshaler,
	ProvideIntegration,
	ProvideKafkaPublisher,
	ProvideRedisClient,
	ProvideWatermillLogger,
	ProvideWorkspaceEvent,
	ProvideWorkspaceEventHubPubSub,
	ProvideWorkspaceEventInternalPubSub,
	wire.Bind(new(app.WorkspaceEventPubSub), new(*WorkspaceEvent)),
)
