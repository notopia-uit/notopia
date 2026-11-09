package persistence

import (
	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgreadmodel"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgrepo"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

var ProviderSet = wire.NewSet(
	ProvideGooseProvider,
	ProvidePg,
	ProvidePgPool,
	ProvideSQLCQueries,
	ProvidePgxPoolStdlib,
	pgrepo.ProvideUnitOfWork,
	pgrepo.ProvideSubscriptionRepo,
	pgrepo.ProvideAccountRepo,
	pgrepo.ProvidePriceRepo,
	pgrepo.ProvidePlanRepo,
	pgrepo.ProvidePolicyRepo,
	pgrepo.ProvideProviderEventLog,
	pgreadmodel.ProvideBillingReadModel,
	wire.Bind(new(pgsqlc.DBTX), new(*pgxpool.Pool)),
	wire.Bind(new(domain.UnitOfWork), new(*pgrepo.UnitOfWork)),
	wire.Bind(new(domain.SubscriptionRepo), new(*pgrepo.SubscriptionRepo)),
	wire.Bind(new(domain.BillingAccountRepo), new(*pgrepo.AccountRepo)),
	wire.Bind(new(domain.PriceRepo), new(*pgrepo.PriceRepo)),
	wire.Bind(new(domain.PlanRepo), new(*pgrepo.PlanRepo)),
	wire.Bind(new(domain.BillingPolicyRepo), new(*pgrepo.PolicyRepo)),
	wire.Bind(new(domain.ProviderEventLog), new(*pgrepo.ProviderEventLog)),
	wire.Bind(new(app.BillingReadModel), new(*pgreadmodel.BillingReadModel)),
)
