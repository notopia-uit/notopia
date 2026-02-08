package persistence

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence/pg"
)

var PostgresProviderSet = wire.NewSet(
	pg.ProvideDBTXPool,
	pg.ProvideQueries,
)
