package app

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideCasbinEnforcer,
	ProvideCreateWorkspaceHandler,
	ProvideDeleteWorkspaceHandler,
	ProvideGetUserWorkspaceItemPermissionsHandler,
	ProvideGetUserWorkspacesHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideHasWorkspaceItemPermissionHandler,
	ProvideHasWorkspacePermissionHandler,
	ProvideUpdateWorkspaceMembersHandler,
	ProvideLeaveWorkspaceHandler,
	wire.Struct(new(App), "*"),
)
