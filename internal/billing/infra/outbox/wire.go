package outbox

import (
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgrepo"
)

// PublisherProviderSet writes events into the outbox; binaries without the forwarder (seed, fixtures) use it alone.
var PublisherProviderSet = wire.NewSet(
	ProvideSchemaAdapter,
	wire.Bind(new(sql.SchemaAdapter), new(*sql.PostgreSQLQueueSchema)),
	ProvidePublisherFactory,
	wire.Bind(new(pgrepo.PublisherFactory), new(*PublisherFactory)),
)

var ProviderSet = wire.NewSet(
	PublisherProviderSet,
	ProvideOffsetsAdapter,
	wire.Bind(new(sql.OffsetsAdapter), new(*sql.PostgreSQLQueueOffsetsAdapter)),
	ProvideForwarder,
)
