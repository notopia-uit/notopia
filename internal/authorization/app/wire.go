package app

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideCreateWorkspaceHandler,
	ProvideUpdateWorkspaceMembersHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideHasWorkspacePermissionHandler,
	ProvideHasWorkspaceItemPermissionHandler,
	ProvideGetUserWorkspaceItemPermissionsHandler,
)
