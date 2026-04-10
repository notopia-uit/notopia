package app

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideCreateWorkspaceHandler,
	ProvideGetUserWorkspaceItemPermissionsHandler,
	ProvideGetUserWorkspacesHandler,
	ProvideGetWorkspaceMembersHandler,
	ProvideHasWorkspaceItemPermissionHandler,
	ProvideHasWorkspacePermissionHandler,
	ProvideUpdateWorkspaceMembersHandler,
)
