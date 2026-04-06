package pubsub

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationPublisher,
	ProvideRedisClient,
	ProvideWorkspaceEventHub,
	wire.Bind(new(app.IntegrationPublisher), new(*IntegrationPublisher)),
	wire.Bind(new(app.WorkspaceEventHub), new(*WorkspaceEventHub)),
)
