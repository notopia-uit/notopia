package authorizationseed

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/internal/authorization/infra"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideSeeder,
	ProvideValidate,
	ProvideViper,
	ProvideConfig,
	wire.FieldsOf(
		new(*Config),
		"Database",
		"General",
		"Log",
	),
	infra.ProviderSetCasbin,
	app.ProvideCasbinEnforcer,
	logging.ProviderSet,
	otel.ProviderSet,
)
