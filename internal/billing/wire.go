package billing

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/controller"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/internal/billing/seed"
	"github.com/notopia-uit/notopia/pkg/common/cli"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	ProvideAppSettings,
	app.ProviderSet,
	component.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	domain.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)

var (
	ServeProviderSet   = ProviderSet
	MigrateProviderSet = wire.NewSet(config.MigrateProviderSet, component.ProvideValidate, logging.ProviderSet, otel.ProviderSet, persistence.ProvidePgPool, persistence.ProvidePgxPoolStdlib, persistence.ProvideGooseProvider, persistence.NewMigrator, wire.Bind(new(cli.Migrator), new(*persistence.Migrator)))
)

var SeedProviderSet = wire.NewSet(seed.ProviderSet, wire.Bind(new(cli.Seeder), new(*seed.Seeder)))
