package persistence

import (
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence/pg"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence/pgsqlc"
)

var PostgresProviderSet = wire.NewSet(
	pg.ProvidePgPool,
	pg.ProvideQueries,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
)
