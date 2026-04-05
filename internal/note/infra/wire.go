package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/common"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence"
	"github.com/notopia-uit/notopia/internal/note/infra/pubsub"
	"github.com/notopia-uit/notopia/internal/note/infra/service"
)

var ProviderSet = wire.NewSet(
	common.ProviderSet,
	service.ProviderSet,
	persistence.PostgresProviderSet,
	pubsub.ProviderSet,
)
