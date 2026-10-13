package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCmds,
	ProvideQueries,

	ProvideRefundRunner,

	ProvideArchivePriceHandler,
	ProvideCompleteCheckoutHandler,
	ProvideCreatePriceHandler,
	ProvideEnsureDefaultsHandler,
	ProvideEnsurePortalConfigurationHandler,
	ProvideSyncRefundHandler,
	ProvideSyncSubscriptionHandler,
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
	ProvideStartCheckoutHandler,
	ProvideStartPortalHandler,
	wire.Struct(new(Server), "*"),
)
