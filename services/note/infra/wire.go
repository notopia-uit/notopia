package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/services/note/infra/persistence"
)

var ProviderSet = wire.NewSet(
	persistence.PostgresProviderSet,
)
