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
	ArchivePrice              ArchivePriceCmd
	CreatePrice               CreatePriceCmd
	EnsureDefaults            EnsureDefaultsCmd
	EnsurePortalConfiguration EnsurePortalConfigurationCmd
	UpdatePlan                UpdatePlanCmd
	UpdatePolicy              UpdatePolicyCmd
}

func NewCmds(
	handlerProvider *HandlerProvider,
	archivePriceHandler *ArchivePriceHandler,
	createPriceHandler *CreatePriceHandler,
	ensureDefaultsHandler *EnsureDefaultsHandler,
	ensurePortalConfigurationHandler *EnsurePortalConfigurationHandler,
	updatePlanHandler *UpdatePlanHandler,
	updatePolicyHandler *UpdatePolicyHandler,
) *Cmds {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Cmds{
		ArchivePrice:              commonhandler.DecorateCmd(hp, archivePriceHandler),
		CreatePrice:               commonhandler.DecorateCmd(hp, createPriceHandler),
		EnsureDefaults:            commonhandler.DecorateCmd(hp, ensureDefaultsHandler),
		EnsurePortalConfiguration: commonhandler.DecorateCmd(hp, ensurePortalConfigurationHandler),
		UpdatePlan:                commonhandler.DecorateCmd(hp, updatePlanHandler),
		UpdatePolicy:              commonhandler.DecorateCmd(hp, updatePolicyHandler),
	}
}

var ProvideCmds = NewCmds

type Queries struct {
	BatchGetEntitlements BatchGetEntitlementsQuery
	GetEntitlements      GetEntitlementsQuery
	GetPlan              GetPlanQuery
	GetPolicy            GetPolicyQuery
	GetPrice             GetPriceQuery
	GetWorkspaceBilling  GetWorkspaceBillingQuery
	GetWorkspacePayer    GetWorkspacePayerQuery
	ListPlans            ListPlansQuery
	ListPrices           ListPricesQuery
}

func NewQueries(
	handlerProvider *HandlerProvider,
	batchGetEntitlementsHandler *BatchGetEntitlementsHandler,
	getEntitlementsHandler *GetEntitlementsHandler,
	getPlanHandler *GetPlanHandler,
	getPolicyHandler *GetPolicyHandler,
	getPriceHandler *GetPriceHandler,
	getWorkspaceBillingHandler *GetWorkspaceBillingHandler,
	getWorkspacePayerHandler *GetWorkspacePayerHandler,
	listPlansHandler *ListPlansHandler,
	listPricesHandler *ListPricesHandler,
) *Queries {
	hp := (*commonhandler.HandlerProvider)(handlerProvider)
	return &Queries{
		BatchGetEntitlements: commonhandler.DecorateQuery(hp, batchGetEntitlementsHandler),
		GetEntitlements:      commonhandler.DecorateQuery(hp, getEntitlementsHandler),
		GetPlan:              commonhandler.DecorateQuery(hp, getPlanHandler),
		GetPolicy:            commonhandler.DecorateQuery(hp, getPolicyHandler),
		GetPrice:             commonhandler.DecorateQuery(hp, getPriceHandler),
		GetWorkspaceBilling:  commonhandler.DecorateQuery(hp, getWorkspaceBillingHandler),
		GetWorkspacePayer:    commonhandler.DecorateQuery(hp, getWorkspacePayerHandler),
		ListPlans:            commonhandler.DecorateQuery(hp, listPlansHandler),
		ListPrices:           commonhandler.DecorateQuery(hp, listPricesHandler),
	}
}

var ProvideQueries = NewQueries

type Server struct {
	Cmds    *Cmds
	Queries *Queries
}
