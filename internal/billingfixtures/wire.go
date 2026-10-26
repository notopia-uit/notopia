package billingfixtures

import (
	"github.com/goforj/wire"
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
	logging.ProviderSet,
	otel.ProviderSet,
)
