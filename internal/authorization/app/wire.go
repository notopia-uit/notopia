package app

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(CreateWorkspaceHandler), "*"),
	wire.Struct(new(UpdateWorkspaceMembersHandler), "*"),
	wire.Struct(new(GetWorkspaceMembersHandler), "*"),
	wire.Struct(new(HasWorkspacePermissionHandler), "*"),
	wire.Struct(new(HasWorkspaceItemPermissionHandler), "*"),
	wire.Struct(new(GetUserWorkspaceItemPermissionsHandler), "*"),
)
