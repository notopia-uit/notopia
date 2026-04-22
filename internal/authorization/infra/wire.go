package infra

import (
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
)

var ProviderSet = wire.NewSet(
	ProvideCasbinAdapter,
	wire.Bind(new(app.CasbinAdapter), new(*gormadapter.Adapter)),
	ProvideGORMDB,
	ProvideIntegrationPublisher,
	wire.Bind(new(app.IntegrationPublisher), new(*IntegrationPublisher)),
)
