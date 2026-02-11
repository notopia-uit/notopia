package logging

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	Provide,
	ProvideStdoutHandler,
)
