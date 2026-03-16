package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/pubsub"
)

var ProviderSet = wire.NewSet(
	persistence.PostgresProviderSet,
	pubsub.ProviderSet,
)
