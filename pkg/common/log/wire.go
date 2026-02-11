package log

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideMulti,
	ProvideStdoutHandler,
)
