package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note/infra/persistence"
)

var ProviderSet = wire.NewSet(
	persistence.PostgresProviderSet,
)
