package persistence

import (
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pg"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

var PostgresProviderSet = wire.NewSet(
	ProvidePg,
	ProvideGooseProvider,
	pg.ProvidePgPool,
	pg.ProvideQueries,
	pg.ProvideStdlib,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
	wire.Bind(new(Persistence), new(*PersistencePg)),
)
