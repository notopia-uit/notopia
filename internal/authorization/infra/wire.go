package infra

import (
	"github.com/casbin/casbin/v3/log"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/pkg/casbin"
)

var ProviderSetCasbinLogger = wire.NewSet(
	ProvideCasbinLogger,
	wire.Bind(new(log.Logger), new(*casbin.SlogLogger)),
)

var ProviderSetCasbin = wire.NewSet(
	ProvideGORMDB,
	ProvideCasbinAdapter,
	wire.Bind(new(app.CasbinAdapter), new(*gormadapter.Adapter)),
	ProviderSetCasbinLogger,
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationPublisher,
	ProviderSetCasbin,
	wire.Bind(new(app.IntegrationPublisher), new(*IntegrationPublisher)),
)
