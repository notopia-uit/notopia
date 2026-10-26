package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCmds,
	ProvideQueries,

	ProvideRefundRunner,
	ProvideCanceller,

	ProvideArchivePriceHandler,
	ProvideCancelNowHandler,
	ProvideCompleteCheckoutHandler,
	ProvideCreatePriceHandler,
	ProvideEnsureDefaultsHandler,
	ProvideEnsurePortalConfigurationHandler,
	ProvideHandlePayerRemovedHandler,
	ProvideHandleWorkspaceDeletedHandler,
	ProvidePruneProviderEventsHandler,
	ProvideReconcileSubscriptionsHandler,
	ProvideSyncRefundHandler,
	ProvideSyncSubscriptionHandler,
	ProvideUpdatePlanHandler,
	ProvideUpdatePolicyHandler,

	ProvideBatchGetEntitlementsHandler,
	ProvideClaimPayerHandler,
	ProvideGetEntitlementsHandler,
	ProvideGetPlanHandler,
	ProvideGetPolicyHandler,
	ProvideGetPriceHandler,
	ProvideGetWorkspaceBillingHandler,
	ProvideGetWorkspacePayerHandler,
	ProvideListPlansHandler,
	ProvideListPricesHandler,
	ProvidePreviewCancelNowHandler,
	ProvideStartCheckoutHandler,
	ProvideStartPortalHandler,
	wire.Struct(new(Server), "*"),
)
