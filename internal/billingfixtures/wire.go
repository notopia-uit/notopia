package billingfixtures

import (
	"github.com/goforj/wire"
	billingconfig "github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/infra/outbox"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideFixtures,
	ProvideValidate,
	ProvideViper,
	ProvideConfig,
	wire.FieldsOf(
		new(*Config),
		"Database",
		"General",
		"Log",
	),
	persistence.ProviderSet,
	outbox.PublisherProviderSet,
	billingconfig.ProvideDefaultDomainEvent,
	logging.ProviderSet,
	otel.ProviderSet,
)
