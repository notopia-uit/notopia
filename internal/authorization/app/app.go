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
	tracer := traceProvider.Tracer("authorization-app")
	return (*HandlerProvider)(
		commonhandler.NewHandlerProvider(
			commonhandler.WithTracer(tracer),
			commonhandler.WithLogger(logger),
		),
	)
}

var ProvideHandlerProvider = NewHandlerProvider

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

	if err := LoadPoliciesFromString(a.Enforcer, PolicyCSV); err != nil {
		slog.ErrorContext(ctx, "BootStrapPolicies: failed to add policies", slog.Any("error", err))
		return fmt.Errorf("failed to bootstrap policies: %w", err)
	}

	slog.InfoContext(ctx, "BootStrapPolicies: permission policies added")
	return nil
}

func (a *App) LoadFixtures(ctx context.Context) error {
	slog.DebugContext(ctx, "LoadFixtures: loading dev fixture policies")

	if err := LoadPoliciesFromString(a.Enforcer, PolicyFixturesCSV); err != nil {
		slog.ErrorContext(ctx, "LoadFixtures: failed to load fixture policies", slog.Any("error", err))
		return fmt.Errorf("failed to load dev fixture policies: %w", err)
	}

	slog.InfoContext(ctx, "LoadFixtures: dev fixture policies loaded")
	return nil
}
