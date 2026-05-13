package app

import (
	"github.com/goforj/wire"
)

var ProviderSetCommand = wire.NewSet(
	ProvideCreateWorkspaceHandler,
	ProvideDeleteWorkspaceHandler,
	ProvideUpdateWorkspaceMembersHandler,
	ProvideLeaveWorkspaceHandler,

	ProvideCmds,
)

var ProviderSetQuery = wire.NewSet(
	ProvideGetUserWorkspaceItemPermissionsHandler,
	ProvideGetUserWorkspacesHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideHasWorkspaceItemPermissionHandler,
	ProvideHasWorkspacePermissionHandler,

	ProvideQueries,
)

var ProviderSet = wire.NewSet(
	ProvideHandlerProvider,
	ProvideCasbinEnforcer,
	ProviderSetCommand,
	ProviderSetQuery,
	wire.Struct(new(App), "*"),
)
