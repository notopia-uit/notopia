package integrationpublisher

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationPublisher,
	wire.Bind(new(app.IntegrationPublisher), new(*IntegrationPublisher)),
)
