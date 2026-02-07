package http

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

var ProviderSet = wire.NewSet(
	ProvideGin,
	ProvideGinSlog,
	ProvideServer,
	ProvideStrictServer,
	wire.Bind(new(note.StrictServerInterface), new(*strictServer)),
)
