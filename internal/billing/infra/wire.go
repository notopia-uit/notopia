package infra

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/infra/notifier"
	stripeadapter "github.com/notopia-uit/notopia/internal/billing/infra/payment/stripe"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/internal/billing/infra/redis"
	"github.com/notopia-uit/notopia/internal/billing/infra/service"
)

var ProviderSet = wire.NewSet(
	persistence.ProviderSet,
	redis.ProvideClient,
	redis.ProvideCronLocker,
	notifier.ProvideKafka,
	service.ProvideAuthorization,
	stripeadapter.Provide,
	wire.Bind(new(app.Notifier), new(*notifier.Kafka)),
	wire.Bind(new(app.AuthorizationSvc), new(*service.Authorization)),
	wire.Bind(new(app.PaymentProvider), new(*stripeadapter.Provider)),
)
