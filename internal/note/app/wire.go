package app

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app/command"
)

var ProviderSet = wire.NewSet(
	command.ProviderSet,
	wire.Struct(new(App), "*"),
)
