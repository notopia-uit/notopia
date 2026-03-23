package integrationevent

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideIntegrationEvent,
)
