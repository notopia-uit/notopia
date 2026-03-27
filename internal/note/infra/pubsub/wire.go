package pubsub

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationMarshaler,
	ProvideIntegrationPubSub,
	ProvideKafkaTracer,
	ProvideRedisClient,
	ProvideWatermillLogger,
	ProvideWorkspaceEvent,
	ProvideWorkspaceEventHubPubSub,
	ProvideWorkspaceEventInternalPubSub,
	wire.Bind(new(pubsub.WorkspaceEvent), new(*WorkspaceEvent)),
)
