package internal

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/notopia-uit/notopia/services/note/internal/app"
	components "github.com/notopia-uit/notopia/services/note/internal/component"
	"github.com/notopia-uit/notopia/services/note/internal/config"
	"github.com/notopia-uit/notopia/services/note/internal/infra"
	controller "github.com/notopia-uit/notopia/services/note/internal/transport"
)

var ProviderSet = wire.NewSet(
	app.ProviderSet,
	components.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)
