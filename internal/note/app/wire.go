package app

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(App), "*"),
)
