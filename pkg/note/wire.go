package note

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/otlp"
	"github.com/notopia-uit/notopia/pkg/note/app"
	"github.com/notopia-uit/notopia/pkg/note/component"
	"github.com/notopia-uit/notopia/pkg/note/config"
	"github.com/notopia-uit/notopia/pkg/note/handler/server"
)

var ProviderSet = wire.NewSet(
	app.ProviderSet,
	components.ProviderSet,
	config.ProviderSet,
	otlp.ProviderSet,
	server.ProvideSet,
)
