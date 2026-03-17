package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/pubsub"
	"github.com/notopia-uit/notopia/internal/note/infra/service"
)

var ProviderSet = wire.NewSet(
	service.ProviderSet,
	persistence.PostgresProviderSet,
	pubsub.ProviderSet,
)
