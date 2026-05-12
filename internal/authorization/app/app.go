package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v3"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

type HandlerProvider commonhandler.HandlerProvider

func NewHandlerProvider(
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
) *HandlerProvider {
	tracer := traceProvider.Tracer("note-app")
	return (*HandlerProvider)(
		commonhandler.NewHandlerProvider(
			commonhandler.WithTracer(tracer),
			commonhandler.WithLogger(logger),
		),
	)
}

var ProvideHandlerProvider = NewHandlerProvider

type (
	CreateWorkspaceCmd        commonhandler.Cmd[CreateWorkspace]
	DeleteWorkspaceCmd        commonhandler.Cmd[DeleteWorkspace]
	UpdateWorkspaceMembersCmd commonhandler.Cmd[UpdateWorkspaceMembers]
	LeaveWorkspaceCmd         commonhandler.Cmd[LeaveWorkspace]
)

type Cmds struct {
	CreateWorkspace        CreateWorkspaceCmd
	DeleteWorkspace        DeleteWorkspaceCmd
	UpdateWorkspaceMembers UpdateWorkspaceMembersCmd
	LeaveWorkspace         LeaveWorkspaceCmd
}

func NewCmds(
	handlerProvider *HandlerProvider,
	createWorkspace *CreateWorkspaceHandler,
	deleteWorkspace *DeleteWorkspaceHandler,
	updateWorkspaceMembers *UpdateWorkspaceMembersHandler,
	leaveWorkspace *LeaveWorkspaceHandler,
) *Cmds {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Cmds{
		CreateWorkspace:        commonhandler.DecorateCmd(hp, createWorkspace),
		DeleteWorkspace:        commonhandler.DecorateCmd(hp, deleteWorkspace),
		UpdateWorkspaceMembers: commonhandler.DecorateCmd(hp, updateWorkspaceMembers),
		LeaveWorkspace:         commonhandler.DecorateCmd(hp, leaveWorkspace),
	}
}

var ProvideCmds = NewCmds

type (
	GetUserWorkspaceItemPermissionsCmd commonhandler.Query[GetUserWorkspaceItemPermissions, WorkspaceItemPermissions]
	GetUserWorkspacesCmd               commonhandler.Query[GetUserWorkspaces, []UserWorkspace]
	GetWorkspaceMembersCmd             commonhandler.Query[GetWorkspaceMembers, []WorkspaceMember]
	HasWorkspaceItemPermissionCmd      commonhandler.Query[HasWorkspaceItemPermission, bool]
	HasWorkspacePermissionCmd          commonhandler.Query[HasWorkspacePermission, bool]
)

type Queries struct {
	GetUserWorkspaceItemPermissions GetUserWorkspaceItemPermissionsCmd
	GetUserWorkspaces               GetUserWorkspacesCmd
	GetWorkspaceMembers             GetWorkspaceMembersCmd
	HasWorkspaceItemPermission      HasWorkspaceItemPermissionCmd
	HasWorkspacePermission          HasWorkspacePermissionCmd
}

func NewQueries(
	handlerProvider *HandlerProvider,
	getUserWorkspaceItemPermissions *GetUserWorkspaceItemPermissionsHandler,
	getUserWorkspaces *GetUserWorkspacesHandler,
	getWorkspaceMembers *GetWorkspaceMembersHandler,
	hasWorkspaceItemPermission *HasWorkspaceItemPermissionHandler,
	hasWorkspacePermission *HasWorkspacePermissionHandler,
) *Queries {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Queries{
		GetUserWorkspaceItemPermissions: commonhandler.DecorateQuery(hp, getUserWorkspaceItemPermissions),
		GetUserWorkspaces:               commonhandler.DecorateQuery(hp, getUserWorkspaces),
		GetWorkspaceMembers:             commonhandler.DecorateQuery(hp, getWorkspaceMembers),
		HasWorkspaceItemPermission:      commonhandler.DecorateQuery(hp, hasWorkspaceItemPermission),
		HasWorkspacePermission:          commonhandler.DecorateQuery(hp, hasWorkspacePermission),
	}
}

var ProvideQueries = NewQueries

type App struct {
	Enforcer *casbin.TransactionalEnforcer
	Cmds     *Cmds
	Queries  *Queries
}

func (a *App) BootStrapPolicies(ctx context.Context) error {
	slog.DebugContext(ctx, "BootStrapPolicies: adding permission policies")

	permissionPolicies := [][]string{
		// Owner permissions on workspace
		{"owner", "workspace", "read"},
		{"owner", "workspace", "edit"},
		{"owner", "workspace", "delete"},
		// Owner permissions on workspace_item
		{"owner", "workspace_item", "read"},
		{"owner", "workspace_item", "write"},
		{"owner", "workspace_item", "delete"},
		// Editor permissions
		{"editor", "workspace", "read"},
		{"editor", "workspace_item", "read"},
		{"editor", "workspace_item", "write"},
		{"editor", "workspace_item", "delete"},
		// Viewer permissions
		{"viewer", "workspace", "read"},
		{"viewer", "workspace_item", "read"},
	}

	_, err := a.Enforcer.AddPolicies(permissionPolicies)
	if err != nil {
		slog.ErrorContext(ctx, "BootStrapPolicies: failed to add permission policies", slog.Any("error", err))
		return fmt.Errorf("failed to add permission policies: %w", err)
	}

	roleInheritancePolicies := [][]string{
		// Role inheritance: note/folder inherit workspace_item (g2 type, 2 params)
		{"note", "workspace_item"},
		{"folder", "workspace_item"},
	}
	_, err = a.Enforcer.AddNamedGroupingPolicies("g2", roleInheritancePolicies)
	if err != nil {
		slog.ErrorContext(ctx, "BootStrapPolicies: failed to add role inheritance policies", slog.Any("error", err))
		return fmt.Errorf("failed to add role inheritance policies: %w", err)
	}

	slog.InfoContext(ctx, "BootStrapPolicies: permission policies added")
	return nil
}

func (a *App) SeedDev(ctx context.Context) error {
	slog.DebugContext(ctx, "SeedDev: seeding dev user workspace policies")

	if err := LoadPoliciesFromString(a.Enforcer, PolicyTestCSV); err != nil {
		slog.ErrorContext(ctx, "SeedDev: failed to load test policies", slog.Any("error", err))
		return fmt.Errorf("failed to seed dev policies: %w", err)
	}

	slog.InfoContext(ctx, "SeedDev: dev user workspace policies seeded")
	return nil
}
