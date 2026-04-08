package outbox

import (
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgrepo"
)

var ProviderSet = wire.NewSet(
	ProvideFromPersistenceToQSLForwarder,
	ProvideSchemaAdapter,
	wire.Bind(new(sql.SchemaAdapter), new(*sql.DefaultPostgreSQLSchema)),
	wire.Bind(new(pgrepo.PublisherFactory), new(*FromPersistenceToQSLForwarder)),

	ProvideOutbox,
)
