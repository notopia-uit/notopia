package authorization

import (
	"github.com/notopia-uit/notopia/internal/authorization/app"
)

type App struct {
	CreateWorkspace                 *app.CreateWorkspaceHandler
	DeleteWorkspace                 *app.DeleteWorkspaceHandler
	GetUserWorkspaceItemPermissions *app.GetUserWorkspaceItemPermissionsHandler
	GetUserWorkspaces               *app.GetUserWorkspacesHandler
	GetWorkspaceMembers             *app.GetWorkspaceMembersHandler
	HasWorkspaceItemPermission      *app.HasWorkspaceItemPermissionHandler
	HasWorkspacePermission          *app.HasWorkspacePermissionHandler
	UpdateWorkspaceMembers          *app.UpdateWorkspaceMembersHandler
}

func (a *App) BootStrapPolicies() {
}

func (a *App) SeedDev() {
}
