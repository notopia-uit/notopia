package app

import (
	"log/slog"

	"github.com/casbin/casbin/v3"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "authorization-app"

type Cmds struct {
	CreateWorkspace        CreateWorkspaceCmd
	DeleteWorkspace        DeleteWorkspaceCmd
	UpdateWorkspaceMembers UpdateWorkspaceMembersCmd
	LeaveWorkspace         LeaveWorkspaceCmd
}

func NewCmds(
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
	createWorkspace *CreateWorkspaceHandler,
	deleteWorkspace *DeleteWorkspaceHandler,
	updateWorkspaceMembers *UpdateWorkspaceMembersHandler,
	leaveWorkspace *LeaveWorkspaceHandler,
) *Cmds {
	d := commonhandler.NewDecorator(
		commonhandler.WithTracer(traceProvider.Tracer(tracerName)),
		commonhandler.WithLogger(logger),
	)
	return &Cmds{
		CreateWorkspace:        d.Cmd(createWorkspace),
		DeleteWorkspace:        d.Cmd(deleteWorkspace),
		UpdateWorkspaceMembers: d.Cmd(updateWorkspaceMembers),
		LeaveWorkspace:         d.Cmd(leaveWorkspace),
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
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
	getUserWorkspaceItemPermissions *GetUserWorkspaceItemPermissionsHandler,
	getUserWorkspaces *GetUserWorkspacesHandler,
	getWorkspaceMembers *GetWorkspaceMembersHandler,
	hasWorkspaceItemPermission *HasWorkspaceItemPermissionHandler,
	hasWorkspacePermission *HasWorkspacePermissionHandler,
) *Queries {
	d := commonhandler.NewDecorator(
		commonhandler.WithTracer(traceProvider.Tracer(tracerName)),
		commonhandler.WithLogger(logger),
	)
	return &Queries{
		GetUserWorkspaceItemPermissions: d.Query(getUserWorkspaceItemPermissions),
		GetUserWorkspaces:               d.Query(getUserWorkspaces),
		GetWorkspaceMembers:             d.Query(getWorkspaceMembers),
		HasWorkspaceItemPermission:      d.Query(hasWorkspaceItemPermission),
		HasWorkspacePermission:          d.Query(hasWorkspacePermission),
	}
}

var ProvideQueries = NewQueries

type App struct {
	Enforcer *casbin.TransactionalEnforcer
	Cmds     *Cmds
	Queries  *Queries
}
