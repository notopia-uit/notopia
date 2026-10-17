package billingfixtures

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence"
	"github.com/notopia-uit/notopia/internal/billingfixtures/config"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideFixtures,
	config.ProviderSet,
	persistence.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)
