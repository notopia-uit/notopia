package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

type HandlerProvider commonhandler.HandlerProvider

func NewHandlerProvider(traceProvider trace.TracerProvider, logger *slog.Logger) *HandlerProvider {
	return (*HandlerProvider)(commonhandler.NewHandlerProvider(
		commonhandler.WithTracer(traceProvider.Tracer("billing-app")),
		commonhandler.WithLogger(logger),
	))
}

var ProvideHandlerProvider = NewHandlerProvider

type Cmds struct {
	EnsureDefaults EnsureDefaultsCmd
}

func NewCmds(
	handlerProvider *HandlerProvider,
	ensureDefaultsHandler *EnsureDefaultsHandler,
) *Cmds {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Cmds{
		EnsureDefaults: commonhandler.DecorateCmd(hp, ensureDefaultsHandler),
	}
}

var ProvideCmds = NewCmds

type Queries struct {
	BatchGetEntitlements BatchGetEntitlementsQuery
	GetEntitlements      GetEntitlementsQuery
	GetWorkspaceBilling  GetWorkspaceBillingQuery
	GetWorkspacePayer    GetWorkspacePayerQuery
	ListPlans            ListPlansQuery
}

func NewQueries(
	handlerProvider *HandlerProvider,
	batchGetEntitlementsHandler *BatchGetEntitlementsHandler,
	getEntitlementsHandler *GetEntitlementsHandler,
	getWorkspaceBillingHandler *GetWorkspaceBillingHandler,
	getWorkspacePayerHandler *GetWorkspacePayerHandler,
	listPlansHandler *ListPlansHandler,
) *Queries {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Queries{
		BatchGetEntitlements: commonhandler.DecorateQuery(hp, batchGetEntitlementsHandler),
		GetEntitlements:      commonhandler.DecorateQuery(hp, getEntitlementsHandler),
		GetWorkspaceBilling:  commonhandler.DecorateQuery(hp, getWorkspaceBillingHandler),
		GetWorkspacePayer:    commonhandler.DecorateQuery(hp, getWorkspacePayerHandler),
		ListPlans:            commonhandler.DecorateQuery(hp, listPlansHandler),
	}
}

var ProvideQueries = NewQueries

type Server struct {
	Cmds    *Cmds
	Queries *Queries
}
