package outbox

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgrepo"
)

var ProviderSet = wire.NewSet(
	ProvideFromPersistenceToQSLForwarder,
	wire.Bind(new(pgrepo.PublisherFactory), new(*FromPersistenceToQSLForwarder)),

	ProvideOutbox,
)
