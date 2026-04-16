package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
)

var ProviderSet = wire.NewSet(
	ProvideCasbinEnforcer,
	ProvideGORMDB,
	ProvideIntegrationPublisher,
	wire.Bind(new(app.IntegrationPublisher), new(*IntegrationPublisher)),
)
