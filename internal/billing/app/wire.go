package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCmds,
	ProvideQueries,

	ProvideArchivePriceHandler,
	ProvideCreatePriceHandler,
	ProvideEnsureDefaultsHandler,
	ProvideEnsurePortalConfigurationHandler,
	ProvideUpdatePlanHandler,
	ProvideUpdatePolicyHandler,

	ProvideBatchGetEntitlementsHandler,
	ProvideGetEntitlementsHandler,
	ProvideGetPlanHandler,
	ProvideGetPolicyHandler,
	ProvideGetPriceHandler,
	ProvideGetWorkspaceBillingHandler,
	ProvideGetWorkspacePayerHandler,
	ProvideListPlansHandler,
	ProvideListPricesHandler,
	wire.Struct(new(Server), "*"),
)
