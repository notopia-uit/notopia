package app

type App struct {
	CreateWorkspace                 *CreateWorkspaceHandler
	DeleteWorkspace                 *DeleteWorkspaceHandler
	GetUserWorkspaceItemPermissions *GetUserWorkspaceItemPermissionsHandler
	GetUserWorkspaces               *GetUserWorkspacesHandler
	GetWorkspaceMembers             *GetWorkspaceMembersHandler
	HasWorkspaceItemPermission      *HasWorkspaceItemPermissionHandler
	HasWorkspacePermission          *HasWorkspacePermissionHandler
	UpdateWorkspaceMembers          *UpdateWorkspaceMembersHandler
}

func (a *App) BootStrapPolicies() {
}

func (a *App) SeedDev() {
}
