package note

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/log"
	"github.com/notopia-uit/notopia/pkg/note/app"
	"github.com/notopia-uit/notopia/pkg/note/component"
	"github.com/notopia-uit/notopia/pkg/note/config"
	"github.com/notopia-uit/notopia/pkg/note/controller"
	"github.com/notopia-uit/notopia/pkg/note/infra"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	app.ProviderSet,
	components.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	infra.ProviderSet,
	log.ProviderSet,
	otel.ProviderSet,
	wire.Value(ServiceName),
	wire.Value(ServiceVersion),
)
