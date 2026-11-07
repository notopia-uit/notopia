package note

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/controller"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/pkg/common/cli"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
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
	MigrateProviderSet = wire.NewSet(
		config.MigrateProviderSet, component.ProvideValidate, logging.ProviderSet, otel.ProviderSet,
		persistence.ProvidePgPool, persistence.ProvidePgxPoolStdlib, persistence.ProvideGooseProvider,
		persistence.NewMigrator, wire.Bind(new(cli.Migrator), new(*persistence.Migrator)),
	)
)
