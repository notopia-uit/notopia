package app

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app/command"
)

var ProviderSetCommand = wire.NewSet(
	wire.Struct(new(command.MoveWorkspaceItemsHandler), "*"),
)

var ProviderSet = wire.NewSet(
	ProviderSetCommand,
	wire.Struct(new(App), "*"),
)
