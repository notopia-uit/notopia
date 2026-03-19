package authorization

import (
	"github.com/notopia-uit/notopia/internal/authorization/app"
)

type App struct {
	CreateWorkspace                 *app.CreateWorkspaceHandler
	UpdateWorkspaceMembers          *app.UpdateWorkspaceMembersHandler
	GetWorkspaceMembers             *app.GetWorkspaceMembersHandler
	HasWorkspacePermission          *app.HasWorkspacePermissionHandler
	HasWorkspaceItemPermission      *app.HasWorkspaceItemPermissionHandler
	GetUserWorkspaceItemPermissions *app.GetUserWorkspaceItemPermissionsHandler
}

func (a *App) BootStrapPolicies() {
}

func (a *App) SeedDev() {
}
