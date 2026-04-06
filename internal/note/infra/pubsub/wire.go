package pubsub

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationPublisher,
	ProvideRedisClient,
	ProvideWorkspaceEvent,
	ProvideWorkspaceEventHubPubSub,
	ProvideWorkspaceEventInternalPubSub,
	wire.Bind(new(app.WorkspaceEventPubSub), new(*WorkspaceEvent)),
)
