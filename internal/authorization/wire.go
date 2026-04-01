package authorization

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/notopia-uit/notopia/pkg/pb"
)

var ProviderSetComponent = wire.NewSet(
	ProvideCasbinEnforcer,
	ProvideGORMDB,
	ProvideValidate,
)

var ProviderSetConfig = wire.NewSet(
	ProvideViper,
	ProvideConfig,
	wire.FieldsOf(
		new(*Config),
		"Database",
		"General",
		"Log",
		"Server",
		"Kafka",
	),
)

var ProviderSetGRPCServer = wire.NewSet(
	ProvideGRPCServiceServer,
	ProvideGRPCServer,
	wire.Bind(new(pb.AuthorizationServiceServer), new(*GRPCServiceServer)),
)

var ProviderSetApp = wire.NewSet(
	app.ProviderSet,
	wire.Struct(new(App), "*"),
)

var ProviderSet = wire.NewSet(
	ProviderSetComponent,
	ProvideServer,
	ProviderSetConfig,
	ProviderSetGRPCServer,
	ProviderSetApp,
	logging.ProviderSet,
	otel.ProviderSet,
)
