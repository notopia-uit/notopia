package app

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCmds,
	ProvideQueries,

	ProvideEnsureDefaultsHandler,

	ProvideBatchGetEntitlementsHandler,
	ProvideGetEntitlementsHandler,
	ProvideGetWorkspaceBillingHandler,
	ProvideGetWorkspacePayerHandler,
	ProvideListPlansHandler,
	wire.Struct(new(Server), "*"),
)
