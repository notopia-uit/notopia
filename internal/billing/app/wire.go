package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideCmds,
	ProvideQueries,

	ProvideSubscriptionCanceller,

	ProvideArchivePriceHandler,
	ProvideCancelNowHandler,
	ProvideCancelProviderSubscriptionHandler,
	ProvideCompleteCheckoutHandler,
	ProvideCreatePriceHandler,
	ProvideEnsureDefaultsHandler,
	ProvideEnsurePortalConfigurationHandler,
	ProvideHandlePayerRemovedHandler,
	ProvideHandleWorkspaceDeletedHandler,
	ProvidePruneProviderEventsHandler,
	ProvideReconcileSubscriptionsHandler,
	ProvideSubmitRefundHandler,
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
