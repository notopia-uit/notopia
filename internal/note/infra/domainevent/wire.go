package domainevent

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideProcessor,
	wire.Bind(new(app.DomainEventProcessor), new(*Processor)),
)
