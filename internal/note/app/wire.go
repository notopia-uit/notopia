package app

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(App), "*"),
	pubsub.ProviderSet,
)
