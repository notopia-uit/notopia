package note

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
	components "github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/infra"
	controller "github.com/notopia-uit/notopia/internal/note/controller"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	app.ProviderSet,
	components.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)
