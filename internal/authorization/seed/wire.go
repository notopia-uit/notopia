package seed

import (
	"github.com/go-playground/validator/v10"
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/internal/authorization/config"
	"github.com/notopia-uit/notopia/internal/authorization/infra"
	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/otel"
)

var ProviderSet = wire.NewSet(
	ProvideSeeder,
	NewValidate,
	config.SeedProviderSet,
	infra.ProviderSetCasbin,
	app.ProvideCasbinEnforcer,
	logging.ProviderSet,
	otel.ProviderSet,
)

func NewValidate() *validator.Validate { return validator.New(validator.WithRequiredStructEnabled()) }
