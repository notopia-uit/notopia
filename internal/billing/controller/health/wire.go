package health

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	Provide,
)
