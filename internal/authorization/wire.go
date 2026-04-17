package authorization

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/internal/authorization/config"
	"github.com/notopia-uit/notopia/internal/authorization/controller/grpc"
	"github.com/notopia-uit/notopia/internal/authorization/infra"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSetComponent = wire.NewSet(
	ProvideValidate,
)

var ProviderSetController = wire.NewSet(
	grpc.ProviderSet,
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	ProviderSetComponent,
	ProviderSetController,
	app.ProviderSet,
	config.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)
