package authorization

import (
	"github.com/notopia-uit/notopia/internal/authorization/app"
)

type App struct {
	CreateWorkspace                 *app.CreateWorkspaceHandler
	GetUserWorkspaceItemPermissions *app.GetUserWorkspaceItemPermissionsHandler
	GetWorkspaceMembers             *app.GetWorkspaceMembersHandler
	GetUserWorkspaces               *app.GetUserWorkspacesHandler
	UpdateWorkspaceMembers          *app.UpdateWorkspaceMembersHandler
	HasWorkspacePermission          *app.HasWorkspacePermissionHandler
	HasWorkspaceItemPermission      *app.HasWorkspaceItemPermissionHandler
}

func (a *App) BootStrapPolicies() {
}

func (a *App) SeedDev() {
}
