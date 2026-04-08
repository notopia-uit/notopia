package workspaceevent

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideRedisClient,
	ProvideWorkspaceEventHub,
	wire.Bind(new(app.WorkspaceEventPublisher), new(*WorkspaceEventHub)),
	wire.Bind(new(app.WorkspaceEventSubscriber), new(*WorkspaceEventHub)),
	wire.Bind(new(app.WorkspaceEventHub), new(*WorkspaceEventHub)),
)
