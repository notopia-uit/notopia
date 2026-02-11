package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/services/note/internal/infra/persistence"
)

var ProviderSet = wire.NewSet(
	persistence.PostgresProviderSet,
)
