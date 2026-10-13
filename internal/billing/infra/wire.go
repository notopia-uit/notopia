package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/infra/notifier"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/internal/billing/infra/redis"
	"github.com/notopia-uit/notopia/internal/billing/infra/service"
)

var ProviderSet = wire.NewSet(
	persistence.ProviderSet,
	redis.ProvideClient,
	notifier.ProvideKafka,
	service.ProvideAuthorization,
	wire.Bind(new(app.Notifier), new(*notifier.Kafka)),
	wire.Bind(new(app.AuthorizationSvc), new(*service.Authorization)),
)
