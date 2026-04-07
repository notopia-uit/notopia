package outbox

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pg"
)

var ProviderSet = wire.NewSet(
	ProvideFromPersistenceToQSLForwarder,
	wire.Bind(new(pg.PublisherFactory), new(*FromPersistenceToQSLForwarder)),

	ProvideOutbox,
)
