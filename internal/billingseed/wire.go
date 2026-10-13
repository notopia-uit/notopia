package billingseed

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	billingconfig "github.com/notopia-uit/notopia/internal/billing/config"
	stripeadapter "github.com/notopia-uit/notopia/internal/billing/infra/payment/stripe"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/internal/billingseed/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

// Billing infra constructors take billing's config types; the small seed config is mapped onto
// them here, so neither config package imports the other or app.

func NewBillingStripeConfig(stripe *config.Stripe) *billingconfig.Stripe {
	return &billingconfig.Stripe{
		SecretKey:     stripe.SecretKey,
		WebhookSecret: "",
	}
}

var ProvideBillingStripeConfig = NewBillingStripeConfig

func NewBillingConfig(catalog *config.Catalog) *billingconfig.Billing {
	return &billingconfig.Billing{
		Currency:   catalog.Currency,
		WebBaseURL: catalog.WebBaseURL,
	}
}

var ProvideBillingConfig = NewBillingConfig

func NewAppSettings(catalog *config.Catalog, identity *commonconfig.Identity) *app.Settings {
	return &app.Settings{
		WebBaseURL:           catalog.WebBaseURL,
		ServiceAccountUserID: identity.ServiceAccountUserID,
	}
}

var ProvideAppSettings = NewAppSettings

var ProviderSet = wire.NewSet(
	ProvideSeeder,
	ProvideBillingStripeConfig,
	ProvideBillingConfig,
	ProvideAppSettings,
	config.ProviderSet,
	persistence.ProviderSet,
	stripeadapter.Provide,
	wire.Bind(new(app.PaymentProvider), new(*stripeadapter.Provider)),
	app.ProvideEnsureDefaultsHandler,
	app.ProvideListPricesHandler,
	app.ProvideCreatePriceHandler,
	app.ProvideEnsurePortalConfigurationHandler,
	logging.ProviderSet,
	otel.ProviderSet,
)
