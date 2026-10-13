package billing

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/controller"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	app.ProviderSet,
	component.ProviderSet,
	config.ProviderSet,
	controller.ProviderSet,
	domain.ProviderSet,
	infra.ProviderSet,
	logging.ProviderSet,
	otel.ProviderSet,
)
