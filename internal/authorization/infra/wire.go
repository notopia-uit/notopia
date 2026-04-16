package infra

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideCasbinEnforcer,
	ProvideGORMDB,
)
